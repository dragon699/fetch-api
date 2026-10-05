"""Serve the last completed summary while refreshing Redis independently."""
import hashlib
import json
from concurrent.futures import ThreadPoolExecutor
from copy import deepcopy
from threading import Lock
from time import time

import requests
from common.telemetry.src.tracing.wrappers import traced
from common.utils.helpers import DataUtils, TimeUtils
from fetch_api.settings import settings
from fetch_api.src.cache.client import RedisClient
from fetch_api.src.client import ConnectorClient
from fetch_api.src.telemetry.logging import log


class SummaryRefresher:
    def __init__(self, cache=None):
        self.cache = cache
        self.executor = None
        self.pending = set()
        self.mutex = Lock()
        self.closed = False

    def start(self):
        with self.mutex:
            self.closed = False
            if self.executor is None:
                self.executor = ThreadPoolExecutor(
                    max_workers=settings.ai_summary_background_workers,
                    thread_name_prefix='ai-summary'
                )

    def close(self):
        with self.mutex:
            self.closed = True
            executor, self.executor = self.executor, None
        if executor:
            executor.shutdown(wait=True, cancel_futures=True)
        with self.mutex:
            self.pending.clear()

    def _cache(self):
        with self.mutex:
            if self.cache is None:
                if any(value is None for value in (
                    settings.redis_host, settings.redis_port, settings.redis_db
                )):
                    raise ValueError('AI summaries require Redis configuration')
                self.cache = RedisClient(
                    settings.redis_host, settings.redis_port, settings.redis_db,
                    settings.redis_password,
                    socket_timeout=settings.ai_summary_cache_timeout_seconds
                )
            return self.cache

    @staticmethod
    def _hash(value):
        return hashlib.sha256(json.dumps(value, sort_keys=True, separators=(',', ':')).encode()).hexdigest()

    @classmethod
    def stream_key(cls, endpoint, upstreams, prompt, template):
        # A stable stream survives telemetry changes, but different query scopes
        # and instruction choices never borrow each other's summaries.
        scope = [{k: u.get(k, {} if k == 'params' else None)
                  for k in ('method', 'endpoint', 'params')} for u in upstreams]
        return 'ai-summary:v1:' + cls._hash({
            'endpoint': endpoint, 'upstreams': scope,
            'prompt': prompt, 'instructions_template': template
        })

    @classmethod
    def result_key(cls, stream, selection):
        return stream + ':result:' + cls._hash(selection)

    @staticmethod
    def _valid_record(record, selection):
        if not isinstance(record, dict) or not isinstance(record.get('item'), dict):
            return False
        item = record['item']
        return (isinstance(item.get('answer'), str) and bool(item['answer'].strip())
                and all(item.get(k) == v for k, v in selection.items())
                and isinstance(record.get('cached_at'), str)
                and isinstance(record.get('fingerprint'), str)
                and isinstance(record.get('refreshed_at'), (int, float)))

    def get_and_refresh(self, endpoint, upstreams, items, prompt, template):
        stream = self.stream_key(endpoint, upstreams, prompt, template)
        fingerprint = self._hash(DataUtils.omit_volatile_data(items))
        summary = None
        try:
            cache = self._cache()
            selection = cache.get(stream + ':selection')
            if isinstance(selection, dict) and selection.get('provider') and selection.get('model'):
                record = cache.get(self.result_key(stream, selection))
                if self._valid_record(record, selection):
                    summary = deepcopy(record['item'])
                    summary.update(
                        cached=True,
                        cached_at=record['cached_at'],
                        stale=(record['fingerprint'] != fingerprint or
                               time() - record['refreshed_at'] >= settings.ai_summary_refresh_interval_seconds)
                    )
        except Exception as err:
            log.warning('AI summary cache unavailable', extra={'error_type': type(err).__name__})
        # No ML HTTP calls take place on this request's critical path.
        self._schedule(stream, fingerprint, deepcopy(items), prompt, template)
        return summary

    def _schedule(self, stream, fingerprint, items, prompt, template):
        with self.mutex:
            # Bound all outstanding work; do not build an unbounded executor queue.
            if self.closed or stream in self.pending or len(self.pending) >= settings.ai_summary_background_workers:
                return
            if self.executor is None:
                self.executor = ThreadPoolExecutor(
                    max_workers=settings.ai_summary_background_workers,
                    thread_name_prefix='ai-summary'
                )
            self.pending.add(stream)
            try:
                self.executor.submit(self._run, stream, fingerprint, items, prompt, template)
            except Exception:
                self.pending.discard(stream)
                raise

    def _run(self, stream, fingerprint, items, prompt, template):
        lock = None
        acquired = False
        try:
            cache = self._cache()
            # Ownership tokens prevent a worker releasing a successor's lock.
            lock = cache.client.lock(
                stream + ':refresh-lock',
                timeout=settings.ai_summary_background_timeout_seconds + 35,
                blocking=False
            )
            acquired = lock.acquire(blocking=False)
            if acquired:
                self._refresh(cache, stream, fingerprint, items, prompt, template)
        except Exception as err:
            log.warning('AI summary background refresh failed', extra={
                'summary_key': stream, 'error_type': type(err).__name__,
                'upstream_status': (err.response.status_code
                                    if isinstance(err, requests.HTTPError) and err.response is not None else None)
            })
        finally:
            if acquired:
                try:
                    lock.release()
                except Exception as err:
                    log.warning('AI summary lock release failed', extra={'error_type': type(err).__name__})
            with self.mutex:
                self.pending.discard(stream)

    @traced('refresh AI summary')
    def _refresh(self, cache, stream, fingerprint, items, prompt, template, span=None):
        start = time()
        client = ConnectorClient('ml', cache=False)
        payload = client.resolve_ml_request('ask', {
            'instructions_template': template,
            'prompt': '{}\n\n\nJSON_DATA: {}'.format(prompt, json.dumps(items, separators=(',', ':')))
        })
        selection = {'provider': payload['provider'], 'model': payload['model']}
        # Resolve the current selection in the worker. Future dashboard requests
        # cannot use an older provider's result once this marker changes.
        cache.set(stream + ':selection', selection, ttl=settings.redis_cache_ttl)
        key = self.result_key(stream, selection)
        existing = cache.get(key)
        if (self._valid_record(existing, selection)
                and existing.get('fingerprint') == fingerprint
                and start - existing.get('refreshed_at', 0) < settings.ai_summary_refresh_interval_seconds):
            return
        response = requests.post(
            client.url + '/ask', headers=client.headers, json=payload,
            timeout=(5, settings.ai_summary_background_timeout_seconds)
        )
        response.raise_for_status()
        result = response.json()
        item = result['items'][0]
        if not isinstance(item, dict) or not isinstance(item.get('answer'), str) or not item['answer'].strip():
            raise ValueError('ML connector returned no summary')
        if item.get('provider') != selection['provider'] or item.get('model') != selection['model']:
            raise ValueError('ML connector returned an unexpected provider or model')
        cache.set(key, {
            'item': item, 'fingerprint': fingerprint,
            'cached_at': TimeUtils.time_now(), 'refreshed_at': time()
        }, ttl=settings.redis_cache_ttl)
        span.set_attributes({'ai.provider': selection['provider'], 'ai.model': selection['model']})
        log.info('AI summary refreshed', extra={
            **selection, 'duration_seconds': round(time() - start, 3)
        })


summary_refresher = SummaryRefresher()
