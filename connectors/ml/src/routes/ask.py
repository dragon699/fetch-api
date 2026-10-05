import requests
from httpx import TimeoutException
from fastapi import APIRouter
from fastapi.responses import JSONResponse
from connectors.ml.src.querier import querier
from connectors.ml.src.schemas.ask import RequestAsk
from connectors.ml.src.telemetry.logging import log


router = APIRouter()


def execute(request: RequestAsk, provider=None) -> JSONResponse:
    try:
        result = querier.commit(
            provider=provider or request.provider,
            prompt=request.prompt,
            model=request.model,
            instructions=request.instructions or '',
            instructions_template=request.instructions_template
        )
        return JSONResponse(content=result, status_code=200)
    except (requests.Timeout, TimeoutException):
        return JSONResponse({'error': 'ML provider timed out'}, status_code=504)
    except Exception as err:
        log.error('ML query failed', extra={'error_type': type(err).__name__})
        return JSONResponse({'error': 'ML provider request failed'}, status_code=502)


@router.post('')
def ask(request: RequestAsk) -> JSONResponse:
    return execute(request)


@router.post('/ollama')
def ask_ollama(request: RequestAsk) -> JSONResponse:
    return execute(request, 'ollama')


@router.post('/openclaw')
def ask_openclaw(request: RequestAsk) -> JSONResponse:
    return execute(request, 'openclaw')
