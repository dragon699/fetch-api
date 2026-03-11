from common.messages.api import client_responses
from connectors.grafana.src.grafana.querier import querier
from connectors.grafana.src.telemetry.logging import log
from common.telemetry.src.tracing.wrappers import traced
from fastapi.responses import JSONResponse



class APIProcessor:
    @staticmethod
    @traced('process request')
    def process_request(
        query_ds_type: str,
        query_id: str,
        query_params: dict | None = None,
        span=None
    ) -> JSONResponse:
        query_params = query_params or {}

        try:
            result = querier.commit(
                query_ds_type=query_ds_type,
                query_id=query_id,
                query_params=query_params
            )

            assert not result is None

            log.info('Query executed successfully', extra={
                'query_ds_type': query_ds_type,
                'query_id': query_id
            })

            return JSONResponse(content=result, status_code=200)
        
        except Exception as err:
            log.error('Query execution failed', extra={
                'query_ds_type': query_ds_type,
                'query_id': query_id,
                'error': str(err)
            })

            return JSONResponse(content=client_responses['server-error'], status_code=500)
