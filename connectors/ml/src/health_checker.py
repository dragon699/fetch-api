from datetime import datetime, timedelta
from connectors.ml.settings import settings
from connectors.ml.src.providers import providers
from connectors.ml.src.telemetry.logging import log


class HealthChecker:
    def __init__(self, scheduler):
        self.scheduler = scheduler
        self.statuses = {}

    def get_next_interval(self):
        return (settings.health_check_interval_seconds if settings.healthy
                else settings.health_retry_interval_seconds)

    def get_status(self):
        previous = settings.healthy
        settings.health_last_check = datetime.now().isoformat().split('.')[0]
        for provider in providers.clients:
            try:
                healthy = providers.ping(provider)
            except Exception as err:
                healthy = False
                log.warning('Provider health check failed', extra={
                    'provider': provider, 'error_type': type(err).__name__
                })
            self.statuses[provider] = {
                'healthy': healthy, 'last_check': settings.health_last_check
            }
        settings.healthy = self.statuses[settings.default_provider]['healthy']
        settings.health_next_check = (
            datetime.now() + timedelta(seconds=self.get_next_interval())
        ).isoformat().split('.')[0]
        if settings.health_job_id and previous != settings.healthy:
            self.update_schedule()

    def create_schedule(self):
        self.get_status()
        self.update_schedule()

    def update_schedule(self):
        self.scheduler.add_job(
            self.get_status, 'interval', seconds=self.get_next_interval(),
            id='health_check_ml', replace_existing=True
        )
        settings.health_job_id = 'health_check_ml'
