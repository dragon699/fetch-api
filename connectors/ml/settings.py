import os
from typing import Any
from pydantic_settings import BaseSettings
from common.utils.helpers import SysUtils
from connectors.ml.src.telemetry.logging import logger
from connectors.ml.src.loaders import SettingsLoader



class Settings(BaseSettings):
    name: str = 'connector-ml'
    listen_host: str = '0.0.0.0'
    listen_port: int = 8070

    url: str

    otel_service_name: str = 'connector-ml'
    otel_service_namespace: str = 'fetch-api'
    otel_service_version: str = SysUtils.get_app_version(f'{os.path.dirname(__file__)}/VERSION')
    otlp_endpoint_grpc: str = 'grafana-alloy.monitoring.svc:4317'

    log_level: str = 'info'
    log_format: str = 'json'

    health_endpoint: str | None = None
    health_job_id: str | None = None
    health_next_check: str | None = None
    health_last_check: str | None = None
    healthy: bool | None = None

    health_check_interval_seconds: int = 180
    health_retry_interval_seconds: int = 15

    instructions_template_path: str = 'connectors/ml/templates/instructions.yaml'

    default_model: str
    default_keep_alive_minutes: int = 15
    default_temperature: float = 0.5


    def model_post_init(self, __context: Any) -> None:
        logger.update_settings(
            log_level=self.log_level,
            log_format=self.log_format
        )

        if self.url.endswith('/'):
            self.url = self.url.rstrip('/')

        if not self.health_endpoint:
            self.health_endpoint = self.url


settings = SettingsLoader.load()
