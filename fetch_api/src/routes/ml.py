from fetch_api.settings import connectors
from fetch_api.src.client import ConnectorClient
from fetch_api.src.api_processor import APIProcessor
from fetch_api.src.schemas.ml import MLBody
from fastapi import APIRouter, Request
from fastapi.responses import JSONResponse


router = APIRouter()
client = ConnectorClient(
    connectors['ml'].name,
    cache=connectors['ml'].cache,
    requests_timeout=connectors['ml'].requests_timeout
)


@router.post('/ask', tags=['connector-ml'], summary='Ask the AI')
def ask_ai(
    request: Request,
    body: MLBody
) -> JSONResponse:
    return APIProcessor.process_request(
        request=request,
        body=body,
        client=client,
        upstreams=[{
            'method': 'POST',
            'endpoint':'ask'
        }]
    )
