from common.messages.api import client_responses
from connectors.ml.src.ollama.querier import querier
from connectors.ml.src.telemetry.logging import log
from connectors.ml.src.feature_flags import is_ai_summary_enabled
from fastapi import APIRouter
from fastapi.responses import JSONResponse
from opentelemetry import trace

from connectors.ml.src.schemas.ollama import RequestAsk


router = APIRouter()


@router.post('/ollama', tags=['ask'], summary='Ask Ollama a question')
def ask_ollama(request: RequestAsk) -> JSONResponse:
    ai_summary_enabled = is_ai_summary_enabled(
        route='/ask/ollama',
        provider='ollama'
    )
    trace.get_current_span().set_attribute('feature.enable_ai_summary', ai_summary_enabled)

    if not ai_summary_enabled and request.instructions_template is not None:
        log.info('Skipping AI summary query because feature flag is disabled', extra={
            'provider': 'ollama',
            'instructions_template': request.instructions_template
        })
        return JSONResponse(content={
            'error': 'Feature flag enable_ai_summary is disabled'
        }, status_code=503)

    try:
        result = querier.commit(
            provider='ollama',
            prompt=request.prompt,
            model=request.model,
            instructions=request.instructions,
            instructions_template=request.instructions_template
        )

        assert not result is None

        log.info('Query executed successfully')

        return JSONResponse(content=result, status_code=200)

    except Exception as err:
        log.error('Query execution failed', extra={
            'error': str(err)
        })

        return JSONResponse(content=client_responses['server-error'], status_code=500)


@router.post('/openclaw', tags=['ask'], summary='Ask OpenClaw a question, or give it a task')
def ask_openclaw(request: RequestAsk) -> JSONResponse:
    ai_summary_enabled = is_ai_summary_enabled(
        route='/ask/openclaw',
        provider='openclaw'
    )
    trace.get_current_span().set_attribute('feature.enable_ai_summary', ai_summary_enabled)

    if not ai_summary_enabled and request.instructions_template is not None:
        log.info('Skipping AI summary query because feature flag is disabled', extra={
            'provider': 'openclaw',
            'instructions_template': request.instructions_template
        })
        return JSONResponse(content={
            'error': 'Feature flag enable_ai_summary is disabled'
        }, status_code=503)

    try:
        result = querier.commit(
            provider='openclaw',
            prompt=request.prompt,
            model=request.model,
            instructions=request.instructions,
            instructions_template=request.instructions_template
        )

        assert not result is None

        log.info('Query executed successfully')

        return JSONResponse(content=result, status_code=200)

    except Exception as err:
        log.error('Query execution failed', extra={
            'error': str(err)
        })

        return JSONResponse(content=client_responses['server-error'], status_code=500)
