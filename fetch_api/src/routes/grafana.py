from fetch_api.settings import connectors
from fetch_api.src.client import ConnectorClient
from fetch_api.src.api_processor import APIProcessor
from fastapi import APIRouter, Request, Query
from fastapi.responses import JSONResponse
from fetch_api.src.schemas.grafana import GrafanaBody


router = APIRouter()
client = ConnectorClient(
    connectors['grafana'].name,
    cache=connectors['grafana'].cache,
    requests_timeout=connectors['grafana'].requests_timeout
)


@router.post('/argocd-apps', tags=['connector-grafana'], summary='Fetch ArgoCD applications and their statuses')
def fetch_argocd_apps(
    request: Request,
    body: GrafanaBody
) -> JSONResponse:
    return APIProcessor.process_request(
        request=request,
        body=body,
        client=client,
        upstreams=[{
            'method': 'GET',
            'endpoint':'prometheus/argocd-apps'
        }],
        ai_prompt='Should I be worried about the state of my ArgoCD applications?',
        ai_instructions_template='default'
    )


@router.post('/car-info', tags=['connector-grafana'], summary='Fetch recent car state and info from Teslamate')
def fetch_car_info(
    request: Request,
    body: GrafanaBody
) -> JSONResponse:
    return APIProcessor.process_request(
        request=request,
        body=body,
        client=client,
        upstreams=[
            {
                'method': 'GET',
                'endpoint': endpoint
            } for endpoint in [
                'postgresql/car-battery',
                'postgresql/car-last-charge',
                'postgresql/car-last-location',
                'postgresql/car-state',
                'postgresql/car-efficiency'
            ]
        ],
        ai_prompt='Give me a summary for my Tesla.',
        ai_instructions_template='default'
    )


@router.post('/car-drives-history', tags=['connector-grafana'], summary='Fetch car drives history from Teslamate')
def fetch_car_drives_history(
    request: Request,
    body: GrafanaBody,
    number_of_drives: int = Query(5, ge=1, le=1000, description='How many recent drive records to fetch')
) -> JSONResponse:
    return APIProcessor.process_request(
        request=request,
        body=body,
        client=client,
        upstreams=[{
            'method': 'GET',
            'endpoint':'postgresql/car-drives-history',
            'params': {
                'number_of_drives': number_of_drives
            }
        }],
        ai_prompt=f'Now analyze the data below for these {number_of_drives} drives.',
        ai_instructions_template='car-drives-history'
    )
