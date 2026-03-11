import os
from fastapi import FastAPI
from apscheduler.schedulers.background import BackgroundScheduler
from collections.abc import AsyncIterator
from contextlib import asynccontextmanager

from common.utils.helpers import SysUtils
from connectors.grafana.settings import settings
from connectors.grafana.src.telemetry.tracing import instrumentor
from connectors.grafana.src.grafana.client import GrafanaClient
from connectors.grafana.src.health_checker import HealthChecker
from connectors.grafana.src.loaders import RoutesLoader


@asynccontextmanager
async def lifespan(app: FastAPI) -> AsyncIterator[None]:
    scheduler.start()
    health_checker.create_schedule()
    grafana_client.authenticate()

    RoutesLoader.load(app, settings)

    yield

    scheduler.shutdown()


scheduler = BackgroundScheduler()
health_checker = HealthChecker(scheduler)
grafana_client = GrafanaClient()

app = FastAPI(
    title='Grafana Connector',
    version=SysUtils.get_app_version(f'{os.path.dirname(__file__)}/../VERSION'),
    description='A connector that runs queries against Grafana data sources.',
    lifespan=lifespan,
    **SysUtils.get_swagger_params()
)

instrumentor.instrument(app)
