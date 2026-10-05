from typing import Any
from fastapi import Request
from fastapi.responses import JSONResponse
from common.telemetry.src.tracing.wrappers import traced
from fetch_api.settings import connectors
from fetch_api.src.telemetry.logging import log
from fetch_api.src.client import ConnectorClient
from fetch_api.src.ai_summary import summary_refresher



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
            response = None
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

                if response.status_code not in (200, 201):
                    raise RuntimeError(f'Upstream returned HTTP {response.status_code}')
                
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
                if client.connector_name == 'ml' and response is not None:
                    upstream['error_status'] = response.status_code
                log.warning('Upstream fetch failed', extra={
                    **common_stream_log_attributes,
                    'error': str(err)
                })

        if client.connector_name != 'ml' and body.ai and results['items']:
            if 'ml' in connectors:
                try:
                    summary = summary_refresher.get_and_refresh(
                        endpoint=request.scope['path'], upstreams=upstreams,
                        items=results['items'], prompt=ai_prompt,
                        template=ai_instructions_template
                    )
                    if summary is not None:
                        results['ai_summary'] = summary
                except Exception as err:
                    log.warning('AI summary scheduling failed', extra={
                        **common_log_attributes, 'error_type': type(err).__name__
                    })
            else:
                log.warning('Skipping AI processing, as ML connector is not enabled', extra=common_log_attributes)

        results['total_items'] = len(results['items'])

        if any(
            upstream['status'] == 'failed'
            for upstream in upstreams
        ):
            status_code = 502
            if client.connector_name == 'ml':
                status_code = next((u['error_status'] for u in upstreams
                                    if u.get('error_status') in (422, 502, 503, 504)), 502)

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
