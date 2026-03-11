from connectors.grafana.settings import settings
from fastapi import APIRouter


router = APIRouter()


@router.get('/health', tags=['internal'], summary='Health check')
def health() -> dict:
    return {
        'connector_name': settings.name,
        'healthy': settings.healthy,
        'health_endpoint': settings.health_endpoint,
        'health_last_check': settings.health_last_check,
        'health_next_check': settings.health_next_check,
        'authenticated': settings.authenticated
    }


@router.get('/ready', tags=['internal'], summary='Readiness check')
def ready() -> dict:
    return {
        'ready': (settings.healthy is True) and (settings.authenticated is True)
    }
