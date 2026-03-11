from typing import Any
from datetime import (datetime, timedelta)
from apscheduler.schedulers.background import BackgroundScheduler
from fetch_api.src.telemetry.logging import log
from fetch_api.settings import settings
from fetch_api.src.client import ConnectorClient

from common.telemetry.src.tracing.wrappers import traced
from common.telemetry.src.tracing.helpers import reword



class HealthChecker:
    def __init__(self, scheduler: BackgroundScheduler, connector: Any) -> None:
        self.scheduler = scheduler
        self.connector = connector


    def get_next_interval(self) -> int:
        return (
            settings.connector_health_check_interval_seconds
            if self.connector.healthy is True
            else settings.connector_health_retry_interval_seconds
        )


    def get_next_check_time(self) -> str:
        ts = datetime.now() + timedelta(seconds=self.get_next_interval())
        return ts.isoformat().split('.')[0]


    @traced('get connector health status')
    def get_connector_status(self, span=None) -> None:
        was_healthy = self.connector.healthy
        self.connector.health_last_check = datetime.now().isoformat().split('.')[0]

        try:
            response = ConnectorClient.ping(
                connector_name=self.connector.name,
                connector_url=self.connector.url,
                health_endpoint=self.connector.health_endpoint
            )

        except Exception as err:
            self.connector.healthy = None

            if was_healthy is not None:
                self.update_connector_schedule()

            else:
                self.connector.health_next_check = self.get_next_check_time()

            log.critical(f'Health check failed, connector is unreachable', extra={
                'connector': self.connector.name,
                'health_endpoint': self.connector.health_endpoint,
                'error': str(err)
            })
            span.set_attributes(
                reword({
                    'health.connector.name': self.connector.name,
                    'health.connector.endpoint': self.connector.health_endpoint,
                    'health.connector.error.message': str(err),
                    'health.connector.error.type': type(err).__name__,
                    'health.connector.last_check': self.connector.health_last_check,
                    'health.connector.next_check': self.connector.health_next_check,
                    'health.connector.status.current': self.connector.healthy,
                    'health.connector.status.previous': was_healthy
                })
            )

            return None
        
        try:
            is_healthy = response.json()['healthy']

        except Exception as err:
            self.connector.healthy = None

            if was_healthy is not None:
                self.update_connector_schedule()

            else:
                self.connector.health_next_check = self.get_next_check_time()

            log.critical(f'Health check failed, got unexpected response', extra={
                'connector': self.connector.name,
                'health_endpoint': self.connector.health_endpoint,
                'error': str(err)
            })
            span.set_attributes(
                reword({
                    'health.connector.name': self.connector.name,
                    'health.connector.endpoint': self.connector.health_endpoint,
                    'health.connector.error.message': str(err),
                    'health.connector.error.type': type(err).__name__,
                    'health.connector.response.content': response.text,
                    'health.connector.response.status_code': response.status_code,
                    'health.connector.last_check': self.connector.health_last_check,
                    'health.connector.next_check': self.connector.health_next_check,
                    'health.connector.status.current': self.connector.healthy,
                    'health.connector.status.previous': was_healthy
                })
            )

            return None

        if isinstance(is_healthy, bool):
            if (response.status_code == 200) and (is_healthy is True):
                self.connector.healthy = True

                if was_healthy is not True:
                    log.debug('Health check successful', extra={
                        'connector': self.connector.name,
                        'health_endpoint': self.connector.health_endpoint
                    })
                    self.update_connector_schedule()

                else:
                    self.connector.health_next_check = self.get_next_check_time()
                
                span.set_attributes(
                    reword({
                        'health.connector.name': self.connector.name,
                        'health.connector.endpoint': self.connector.health_endpoint,
                        'health.connector.last_check': self.connector.health_last_check,
                        'health.connector.next_check': self.connector.health_next_check,
                        'health.connector.status.current': self.connector.healthy,
                        'health.connector.status.previous': was_healthy
                    })
                )

            else:
                self.connector.healthy = False

                if was_healthy is not False:
                    self.update_connector_schedule()

                else:
                    self.connector.health_next_check = self.get_next_check_time()
                
                log.critical('Health check failed', extra={
                    'connector': self.connector.name,
                    'health_endpoint': self.connector.health_endpoint,
                    'response_status_code': response.status_code
                })
                span.set_attributes(
                    reword({
                        'health.connector.name': self.connector.name,
                        'health.connector.endpoint': self.connector.health_endpoint,
                        'health.connector.response.content': response.text,
                        'health.connector.response.status_code': response.status_code,
                        'health.connector.last_check': self.connector.health_last_check,
                        'health.connector.next_check': self.connector.health_next_check,
                        'health.connector.status.current': self.connector.healthy,
                        'health.connector.status.previous': was_healthy
                    })
                )

        elif is_healthy is None:
            self.connector.healthy = None

            if was_healthy is not None:
                self.update_connector_schedule()

            else:
                self.connector.health_next_check = self.get_next_check_time()

            log.critical(f'Health check failed, got unexpected response', extra={
                'connector': self.connector.name,
                'health_endpoint': self.connector.health_endpoint,
                'response_status_code': response.status_code
            })
            span.set_attributes(
                reword({
                    'health.connector.name': self.connector.name,
                    'health.connector.endpoint': self.connector.health_endpoint,
                    'health.connector.response.content': response.text,
                    'health.connector.response.status_code': response.status_code,
                    'health.connector.last_check': self.connector.health_last_check,
                    'health.connector.next_check': self.connector.health_next_check,
                    'health.connector.status.current': self.connector.healthy,
                    'health.connector.status.previous': was_healthy
                })
            )


    @traced('create health check schedule')
    def create_connector_schedule(self, span=None) -> None:
        log.debug(f'Scheduling health checks for {self.connector.name}')
        self.get_connector_status()

        if self.connector.health_job_id is None:
            self.update_connector_schedule()


    @traced('update health check schedule')
    def update_connector_schedule(self, span=None) -> None:
        if not self.connector.health_job_id is None:
            try:
                self.scheduler.remove_job(self.connector.health_job_id)

            except:
                pass

        interval_seconds = self.get_next_interval()
        self.connector.health_next_check = self.get_next_check_time()

        job = self.scheduler.add_job(
            self.get_connector_status,
            'interval',
            seconds=interval_seconds,
            id=f'health_check_{self.connector.name}'
        )

        self.connector.health_job_id = job.id
        
        span.set_attributes(
            reword({
                'connector.name': self.connector.name,
                'connector.scheduler.interval.seconds': interval_seconds,
                'connector.scheduler.job.id': self.connector.health_job_id,
                'health.connector.last_check': self.connector.health_last_check,
                'health.connector.next_check': self.connector.health_next_check
            })
        )
