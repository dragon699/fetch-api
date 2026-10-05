import os
from typing import Any, Literal
from pydantic_settings import BaseSettings
from pydantic import Field
from common.utils.helpers import SysUtils
from connectors.ml.src.telemetry.logging import logger
from connectors.ml.src.loaders import SettingsLoader



class Settings(BaseSettings):
    name: str = 'connector-ml'
    listen_host: str = '0.0.0.0'
    listen_port: int = 8070

    url: str | None = None

    otel_service_name: str = 'connector-ml'
    otel_service_namespace: str = 'fetch-api'
    otel_service_version: str = SysUtils.get_app_version(f'{os.path.dirname(__file__)}/VERSION')
    otlp_endpoint_http: str = 'http://grafana-alloy.monitoring.svc:4318'

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

    default_provider: Literal['ollama', 'openclaw'] = 'ollama'
    default_model: str | None = None
    ollama_url: str | None = None
    ollama_default_model: str | None = None
    ollama_timeout_seconds: float = Field(default=90, gt=0)
    openclaw_url: str | None = None
    openclaw_token: str | None = None
    openclaw_default_model: str = 'openclaw/default'
    openclaw_timeout_seconds: float = Field(default=90, gt=0)
    default_keep_alive_minutes: int = 15
    default_temperature: float = 0.5


    def model_post_init(self, __context: Any) -> None:
        if self.default_provider == 'ollama':
            if not (self.ollama_url or self.url) or not (self.ollama_default_model or self.default_model):
                raise ValueError('Ollama requires OLLAMA_URL (or URL) and OLLAMA_DEFAULT_MODEL (or DEFAULT_MODEL)')
        elif not self.openclaw_url or not self.openclaw_token:
            raise ValueError('OpenClaw requires OPENCLAW_URL and OPENCLAW_TOKEN')

        logger.update_settings(
            log_level=self.log_level,
            log_format=self.log_format
        )

        self.url = self.ollama_url or self.url
        if self.url and self.url.endswith('/'):
            self.url = self.url.rstrip('/')

        if not self.health_endpoint:
            self.health_endpoint = self.url


settings = SettingsLoader.load()
