import os
from fastapi import FastAPI
from apscheduler.schedulers.background import BackgroundScheduler
from collections.abc import AsyncIterator
from contextlib import asynccontextmanager

from common.utils.helpers import SysUtils
from connectors.ml.settings import settings
from connectors.ml.src.telemetry.tracing import instrumentor
from connectors.ml.src.ollama.client import OllamaClient
from connectors.ml.src.health_checker import HealthChecker
from connectors.ml.src.loaders import RoutesLoader


@asynccontextmanager
async def lifespan(app: FastAPI) -> AsyncIterator[None]:
    scheduler.start()
    health_checker.create_schedule()

    RoutesLoader.load(app, settings)

    yield

    scheduler.shutdown()


scheduler = BackgroundScheduler()
health_checker = HealthChecker(scheduler)
ollama_client = OllamaClient()

app = FastAPI(
    title='ML Connector',
    version=SysUtils.get_app_version(f'{os.path.dirname(__file__)}/../VERSION'),
    description='A connector that runs LLM queries against Ollama models.',
    lifespan=lifespan,
    **SysUtils.get_swagger_params()
)

instrumentor.instrument(app)
