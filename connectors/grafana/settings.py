import os
from typing import Any
from pydantic_settings import BaseSettings
from common.utils.helpers import SysUtils
from connectors.grafana.src.telemetry.logging import logger
from connectors.grafana.src.loaders import SettingsLoader



class Settings(BaseSettings):
    name: str = 'connector-grafana'
    listen_host: str = '0.0.0.0'
    listen_port: int = 8080

    url: str
    sa_token: str

    otel_service_name: str = 'connector-grafana'
    otel_service_namespace: str = 'fetch-api'
    otel_service_version: str = SysUtils.get_app_version(f'{os.path.dirname(__file__)}/VERSION')
    otlp_endpoint_grpc: str = 'grafana-alloy.monitoring.svc:4317'

    log_level: str = 'info'
    log_format: str = 'json'

    auth_endpoint: str | None = None
    authenticated: bool = False

    health_endpoint: str | None = None
    health_job_id: str | None = None
    health_next_check: str | None = None
    health_last_check: str | None = None
    healthy: bool | None = None

    health_check_interval_seconds: int = 180
    health_retry_interval_seconds: int = 15

    querier_templates_dir: str = 'connectors/grafana/templates'
    querier_default_endpoint: str = 'api/ds/query'
    querier_ds_uid_postgresql: str = 'internal-teslamate'
    querier_ds_uid_prometheus: str = 'internal-victoriametrics'


    def model_post_init(self, __context: Any) -> None:
        logger.update_settings(
            log_level=self.log_level,
            log_format=self.log_format
        )

        if self.url.endswith('/'):
            self.url = self.url.rstrip('/')

        if not self.health_endpoint:
            self.health_endpoint = f'{self.url}/api/health'

        if not self.auth_endpoint:
            self.auth_endpoint = f'{self.url}/api/org'


settings = SettingsLoader.load()
