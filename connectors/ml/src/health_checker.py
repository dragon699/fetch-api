from datetime import (datetime, timedelta)
from apscheduler.schedulers.background import BackgroundScheduler
from connectors.ml.src.telemetry.logging import log
from connectors.ml.settings import settings

from common.telemetry.src.tracing.wrappers import traced
from common.telemetry.src.tracing.helpers import reword
from connectors.ml.src.ollama.client import OllamaClient



class HealthChecker:
    def __init__(self, scheduler: BackgroundScheduler) -> None:
        self.scheduler = scheduler


    def get_next_interval(self) -> int:
        return (
            settings.health_check_interval_seconds
            if settings.healthy is True
            else settings.health_retry_interval_seconds
        )


    def get_next_check_time(self) -> str:
        ts = datetime.now() + timedelta(seconds=self.get_next_interval())
        return ts.isoformat().split('.')[0]


    @traced('get health status')
    def get_status(self, span=None) -> None:
        was_healthy = settings.healthy
        settings.health_last_check = datetime.now().isoformat().split('.')[0]

        try:
            response = OllamaClient.ping()

        except Exception as err:
            settings.healthy = None

            if was_healthy is not None:
                self.update_schedule()

            else:
                settings.health_next_check = self.get_next_check_time()

            log.critical(f'Health check failed, Ollama is unreachable', extra={
                'health_endpoint': settings.health_endpoint,
                'error': str(err)
            })
            span.set_attributes(
                reword({
                    'health.endpoint': settings.health_endpoint,
                    'health.error.message': str(err),
                    'health.error.type': type(err).__name__,
                    'health.last_check': settings.health_last_check,
                    'health.next_check': settings.health_next_check,
                    'health.status.current': settings.healthy,
                    'health.status.previous': was_healthy
                })
            )

            return None

        try:
            is_healthy = response.text == 'Ollama is running'

        except Exception as err:
            settings.healthy = None

            if was_healthy is not None:
                self.update_schedule()

            else:
                settings.health_next_check = self.get_next_check_time()

            log.critical(f'Health check failed, got unexpected response', extra={
                'health_endpoint': settings.health_endpoint,
                'error': str(err)
            })
            span.set_attributes(
                reword({
                    'health.endpoint': settings.health_endpoint,
                    'health.error.message': str(err),
                    'health.error.type': type(err).__name__,
                    'health.response.content': response.text,
                    'health.response.status_code': response.status_code,
                    'health.last_check': settings.health_last_check,
                    'health.next_check': settings.health_next_check,
                    'health.status.current': settings.healthy,
                    'health.status.previous': was_healthy
                })
            )

            return None

        if (response.status_code == 200) and (is_healthy is True):
            settings.healthy = True

            if was_healthy is not True:
                log.debug('Health check successful', extra={
                    'health_endpoint': settings.health_endpoint
                })
                self.update_schedule()

            else:
                settings.health_next_check = self.get_next_check_time()

            span.set_attributes(
                reword({
                    'health.endpoint': settings.health_endpoint,
                    'health.last_check': settings.health_last_check,
                    'health.next_check': settings.health_next_check,
                    'health.status.current': settings.healthy,
                    'health.status.previous': was_healthy
                })
            )

        else:
            settings.healthy = False

            if was_healthy is not False:
                self.update_schedule()
            
            else:
                settings.health_next_check = self.get_next_check_time()

            log.critical('Health check failed, Ollama is unhealthy', extra={
                'health_endpoint': settings.health_endpoint,
                'response_status_code': response.status_code
            })
            span.set_attributes(
                reword({
                    'health.endpoint': settings.health_endpoint,
                    'health.response.content': response.text,
                    'health.response.status_code': response.status_code,
                    'health.last_check': settings.health_last_check,
                    'health.next_check': settings.health_next_check,
                    'health.status.current': settings.healthy,
                    'health.status.previous': was_healthy
                })
            )


    @traced('schedule health checks')
    def create_schedule(self, span=None) -> None:
        log.debug(f'Scheduling health checks')
        self.get_status()

        if settings.health_job_id is None:
            self.update_schedule()


    @traced('update health check schedule')
    def update_schedule(self, span=None) -> None:
        if not settings.health_job_id is None:
            try:
                self.scheduler.remove_job(settings.health_job_id)

            except:
                pass

        interval_seconds = self.get_next_interval()
        settings.health_next_check = self.get_next_check_time()

        job = self.scheduler.add_job(
            self.get_status,
            'interval',
            seconds=interval_seconds,
            id=f'health_check_ollama'
        )

        settings.health_job_id = job.id

        span.set_attributes(
            reword({
                'scheduler.interval.seconds': interval_seconds,
                'scheduler.job.id': settings.health_job_id,
                'health.last_check': settings.health_last_check,
                'health.next_check': settings.health_next_check
            })
        )
