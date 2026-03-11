import json
from typing import Any
from fastapi import Request
from fastapi.responses import JSONResponse
from common.telemetry.src.tracing.wrappers import traced
from common.utils.helpers import DataUtils
from fetch_api.settings import connectors, settings
from fetch_api.src.telemetry.logging import log
from fetch_api.src.client import ConnectorClient



class APIProcessor:
    @staticmethod
    @traced('process request')
    def process_request(
        request: Request,
        body: Any,
        client: ConnectorClient,
        upstreams: list[dict[str, Any]],
        ai_prompt: str | None = None,
        ai_instructions_template: str = 'default',
        span=None
    ) -> JSONResponse:
        status_code = None
        results = {'total_items': 0, 'items': []}
        common_log_attributes = {
            'connector': client.connector_name,
            'endpoint': request.scope['path']
        }

        for upstream in upstreams:
            common_stream_log_attributes = {
                **common_log_attributes.copy(),
                'upstream_endpoint': upstream['endpoint']
            }

            try:
                upstream_method = upstream['method']
                upstream_endpoint = upstream['endpoint']
                params = upstream.get('params', {})

                if upstream_method == 'GET':
                    response = client.get(
                        endpoint=upstream_endpoint,
                        params=params
                    )

                elif upstream_method == 'POST':
                    response = client.post(
                        endpoint=upstream_endpoint,
                        params=params,
                        data=body.model_dump(
                            exclude={'ai'}
                        )
                    )

                assert response.status_code in (200, 201)
                
                response_body = response.json()
                results['items'].extend(response_body['items'])
                upstream['status'] = 'success'
                
                if response_body.get('cached') and response_body['cached'] is True:
                    common_stream_log_attributes['cache_status'] = 'hit'
                    upstream['cache'] = {
                        'cached': True,
                        'cached_at': response_body['cached_at']
                    }

                log.debug('Upstream fetch completed', extra=common_stream_log_attributes)

            except Exception as err:
                upstream['status'] = 'failed'
                log.warning('Upstream fetch failed', extra={
                    **common_stream_log_attributes,
                    'error': str(err)
                })

        if client.connector_name != 'ml':
            if body.ai and len(results['items']) > 0:
                if 'ml' in connectors:
                    ml_client = ConnectorClient(
                        connectors['ml'].name,
                        cache=True,
                        requests_timeout=settings.ai_summary_requests_timeout
                    )

                    upstream_ml_endpoint = 'ask/ollama'
                    commong_ml_log_attributes = {
                        **common_log_attributes,
                        'upstream_ml_endpoint': upstream_ml_endpoint
                    }

                    try:
                        response = ml_client.post(
                            endpoint=upstream_ml_endpoint,
                            data={
                                'instructions_template': ai_instructions_template,
                                'prompt': '{}\n\n\nJSON_DATA: {}'.format(
                                    ai_prompt,
                                    json.dumps(
                                        results['items'],
                                        separators=(',', ':')
                                    )
                                )
                            },
                            cache_key=DataUtils.create_cache_key(
                                connector_name=ml_client.connector_name,
                                method='POST',
                                endpoint=upstream_ml_endpoint,
                                params={},
                                data=DataUtils.omit_volatile_data(results['items'])
                            )
                        )

                        assert response.status_code == 200

                        response_body = response.json()
                        results['ai_summary'] = response_body['items'][0]

                        if response_body.get('cached') and response_body['cached'] is True:
                            commong_ml_log_attributes['cache_status'] = 'hit'
                            results['ai_summary']['cached'] = True
                            results['ai_summary']['cached_at'] = response_body['cached_at']

                        else:
                            commong_ml_log_attributes['cache_status'] = 'miss'

                        log.debug('Fetched AI summary for upstream responses', extra=commong_ml_log_attributes)

                    except Exception as err:
                        log.warning('AI summary fetch failed', extra={
                            **commong_ml_log_attributes,
                            'error': str(err)
                        })

                else:
                    log.warning('Skipping AI processing, as ML connector is not enabled', extra=common_log_attributes)

        results['total_items'] = len(results['items'])

        if any(
            upstream['status'] == 'failed'
            for upstream in upstreams
        ):
            status_code = 502

        else:
            status_code = 200

        for upstream in upstreams:
            if 'cache' in upstream:
                if 'cache' not in results:
                    results['cache'] = []

                results['cache'].append({
                    'upstream_endpoint': upstream['endpoint'],
                    **upstream['cache']
                })

        log.info('Fetch completed', extra=common_log_attributes)

        return JSONResponse(
            status_code=status_code,
            content=results
        )
