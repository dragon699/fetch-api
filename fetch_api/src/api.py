import os
from fastapi import FastAPI
from apscheduler.schedulers.background import BackgroundScheduler
from collections.abc import AsyncIterator
from contextlib import asynccontextmanager

from common.utils.helpers import SysUtils
from fetch_api.settings import (settings, connectors)
from fetch_api.src.telemetry.tracing import instrumentor
from fetch_api.src.health_checker import HealthChecker
from fetch_api.src.loaders import RoutesLoader


@asynccontextmanager
async def lifespan(app: FastAPI) -> AsyncIterator[None]:
    scheduler.start()

    for connector_name in connectors:
        health_checkers[connector_name] = HealthChecker(
            scheduler,
            connectors[connector_name]
        )
        health_checkers[connector_name].create_connector_schedule()

    RoutesLoader.load(app, settings)

    yield

    scheduler.shutdown()


scheduler = BackgroundScheduler()
health_checkers = {}

app = FastAPI(
    title='FetchAPI',
    version=SysUtils.get_app_version(f'{os.path.dirname(__file__)}/../VERSION'),
    description='An API Gateway that fetches data from upstream connectors such as grafana, ollama, kubernetes etc.',
    lifespan=lifespan,
    **SysUtils.get_swagger_params()
)

instrumentor.instrument(app)
