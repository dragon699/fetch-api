from common.messages.api import client_responses
from connectors.ml.src.ollama.querier import querier
from connectors.ml.src.telemetry.logging import log
from fastapi import APIRouter
from fastapi.responses import JSONResponse

from connectors.ml.src.schemas.ollama import RequestAsk


router = APIRouter()


@router.post('/ollama', tags=['ask'], summary='Ask Ollama a question')
def ask_ollama(request: RequestAsk) -> JSONResponse:
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
