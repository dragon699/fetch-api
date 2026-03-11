import requests
from requests import exceptions as ReqExceptions
from connectors.grafana.settings import settings
from common.telemetry.src.tracing.wrappers import traced
from common.telemetry.src.tracing.helpers import reword
from connectors.grafana.src.telemetry.logging import log



class GrafanaClient:
    def __init__(self) -> None:
        self.url = settings.url
        self.sa_token = settings.sa_token
        self.set_headers()


    def set_headers(self) -> None:
        self.headers = {
            'Authorization': f'Bearer {self.sa_token}',
            'Content-Type': 'application/json',
            'Accept': 'application/json'
        }


    @staticmethod
    @traced('ping grafana')
    def ping(span=None) -> requests.Response:
        span.set_attributes({
            'grafana.operation': 'ping',
            'grafana.url': settings.url,
            'grafana.endpoint': settings.health_endpoint
        })

        response = requests.get(
            settings.health_endpoint,
            timeout=5
        )
        
        return response


    @traced('authenticate')
    def authenticate(self, span=None) -> bool:
        if not settings.healthy:
            settings.authenticated = False
            return False

        try:
            response = requests.get(
                settings.auth_endpoint,
                headers=self.headers
            )

            if response.status_code == 200:
                if not settings.authenticated:
                    log.debug(f'Authentication successful', extra={
                        'auth_endpoint': settings.auth_endpoint
                    })

                    settings.authenticated = True

                span.set_attributes(
                    reword({
                        'grafana.operation': 'authentication',
                        'grafana.method': 'GET',
                        'grafana.response.status_code': response.status_code,
                        'grafana.url': settings.url,
                        'grafana.endpoint': settings.auth_endpoint,
                        'grafana.auth.status': settings.authenticated
                    })
                )

                return True

            else:
                settings.authenticated = False

                log.critical(f'Authentication failed', extra={
                    'auth_endpoint': settings.auth_endpoint,
                    'response_status_code': response.status_code
                })
                span.set_attributes(
                    reword({
                        'grafana.operation': 'authentication',
                        'grafana.method': 'GET',
                        'grafana.response.content': response.text,
                        'grafana.response.status_code': response.status_code,
                        'grafana.url': settings.url,
                        'grafana.endpoint': settings.auth_endpoint,
                        'grafana.auth.status': settings.authenticated
                    })
                )

                return False

        except ReqExceptions.RequestException as err:
            settings.authenticated = False

            log.critical(f'Authentication failed, Grafana is unreachable', extra={
                'auth_endpoint': settings.auth_endpoint,
                'error': str(err)
            })
            span.set_attributes(
                reword({
                    'grafana.operation': 'authentication',
                    'grafana.method': 'GET',
                    'grafana.error.message': str(err),
                    'grafana.error.type': type(err).__name__,
                    'grafana.url': settings.url,
                    'grafana.endpoint': settings.auth_endpoint,
                    'grafana.auth.status': settings.authenticated
                })
            )

            return False


    @traced('GET /:grafana')
    def get(self, endpoint: str, data: dict = {}, span=None) -> requests.Response | None:
        if self.authenticate():
            span.set_attributes(
                reword({
                    'grafana.operation': 'get_request',
                    'grafana.method': 'GET',
                    'grafana.request.body': data,
                    'grafana.url': settings.url,
                    'grafana.endpoint': endpoint,
                    'grafana.auth.status': settings.authenticated
                })
            )

            response = requests.get(
                f'{self.url}/{endpoint}',
                headers=self.headers,
                json=data
            )

            return response
        
        else:
            span.set_attributes(
                reword({
                    'grafana.auth.status': settings.authenticated
                })
            )


    @traced('POST /:grafana')
    def post(self, endpoint: str, data: dict = {}, span=None) -> requests.Response | None:
        if self.authenticate():
            span.set_attributes(
                reword({
                    'grafana.operation': 'post_request',
                    'grafana.method': 'POST',
                    'grafana.request.body': data,
                    'grafana.url': settings.url,
                    'grafana.endpoint': endpoint,
                    'grafana.auth.status': settings.authenticated
                })
            )

            response = requests.post(
                f'{self.url}/{endpoint}',
                headers=self.headers,
                json=data
            )

            return response

        else:
            span.set_attributes(
                reword({
                    'grafana.auth.status': settings.authenticated
                })
            )
