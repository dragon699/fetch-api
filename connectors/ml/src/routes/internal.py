from connectors.ml.settings import settings
from fastapi import APIRouter
from connectors.ml.src.providers import providers


router = APIRouter()


@router.get('/health', tags=['internal'], summary='Health check')
def health() -> dict:
    from connectors.ml.src.api import health_checker
    return {
        'default_provider': settings.default_provider,
        'providers': health_checker.statuses,
        'connector_name': settings.name,
        'healthy': settings.healthy,
        'health_endpoint': settings.health_endpoint,
        'health_last_check': settings.health_last_check,
        'health_next_check': settings.health_next_check
    }


@router.get('/ready', tags=['internal'], summary='Readiness check')
def ready() -> dict:
    return {
        'ready': settings.healthy
    }


@router.get('/config', tags=['internal'], summary='ML provider defaults')
def configuration() -> dict:
    return providers.configuration()
