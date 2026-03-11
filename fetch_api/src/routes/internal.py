from fetch_api.settings import (settings, connectors)
from fastapi import APIRouter


router = APIRouter()


@router.get('/health', tags=['internal'], summary='Health check')
def health() -> dict:
    return {
        'healthy': True,
        'connectors': [
            {
                'name': connectors[connector_name].name,
                'healthy': connectors[connector_name].healthy,
                'last_check': connectors[connector_name].health_last_check,
                'next_check': connectors[connector_name].health_next_check,
                'health_endpoint': connectors[connector_name].health_endpoint
            } for connector_name in connectors
        ]
    }


@router.get('/settings', tags=['internal'], summary='Show fetch-api settings')
def read_settings() -> dict:
    return {'settings': settings.model_dump()}


@router.get('/connectors', tags=['internal'], summary='Show connectors` settings')
def read_connectors() -> dict:
    return {'connectors': connectors}


@router.get('/connectors/{connector_name}', tags=['internal'], summary='Show specific connector settings')
def read_connector(connector_name: str) -> dict:
    try:
        return connectors[connector_name].model_dump()

    except:
        return {'error': 'Connector not found'}
