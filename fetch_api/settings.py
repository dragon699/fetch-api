import os
from typing import Any
from pydantic_settings import BaseSettings
from common.utils.helpers import SysUtils
from fetch_api.src.telemetry.logging import logger
from fetch_api.src.loaders import SettingsLoader



class FetchAPISettings(BaseSettings):
    name: str = 'fetch-api'
    listen_host: str = '0.0.0.0'
    listen_port: int = 8079
    listen_url: str | None = None

    otel_service_name: str = 'fetch-api'
    otel_service_namespace: str = 'fetch-api'
    otel_service_version: str = SysUtils.get_app_version(f'{os.path.dirname(__file__)}/VERSION')
    otlp_endpoint_grpc: str = 'grafana-alloy.monitoring.svc:4317'

    redis_host: str | None = None
    redis_port: int | None = None
    redis_db: int | None = None
    redis_password: str | None = ''
    redis_cache_ttl: int = 86400

    log_level: str = 'info'
    log_format: str = 'json'

    connectors: str

    connector_health_check_interval_seconds: int = 20
    connector_health_retry_interval_seconds: int = 5

    ai_summary_requests_timeout: int = 4


    def model_post_init(self, __context: Any) -> None:
        if not self.listen_url:
            self.listen_url = f'http://{self.listen_host}:{self.listen_port}'

        logger.update_settings(
            log_level=self.log_level,
            log_format=self.log_format
        )


class ConnectorSettings(BaseSettings):
    name: str
    host: str
    port: int
    protocol: str = 'http'
    url: str | None = None

    health_endpoint: str | None = None
    health_job_id: str | None = None
    health_next_check: str | None = None
    health_last_check: str | None = None
    healthy: bool | None = None

    cache: bool = False
    requests_timeout: int = 5


    def model_post_init(self, __context: Any) -> None:
        if not self.url:
            self.url = f'{self.protocol}://{self.host}:{self.port}'

        if not self.health_endpoint:
            self.health_endpoint = f'{self.url}/api/health'


settings, connectors = SettingsLoader().load()
