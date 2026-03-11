import json
from pydantic import ValidationError
from fetch_api.src.telemetry.logging import log



class SettingsLoader:
    @staticmethod
    def load():
        from fetch_api.settings import (FetchAPISettings, ConnectorSettings)

        try:
            settings = FetchAPISettings()

        except ValidationError as err:
            print(json.dumps([
                {
                    'error': error['msg'],
                    'env': f'{error["loc"][0].upper()}'
                } for error in err.errors()
            ], indent=2))
            raise SystemExit(1)

        connectors = {}
        connectors_list = [
            conn_name for conn_name in settings.connectors.split(',')
            if conn_name
        ]

        if len(connectors_list) == 0:
            log.critical('Invalid connectors list supplied', extra={
                'errors': [
                    {
                        'error': 'CONNECTORS should be a comma-separated string with connector names',
                        'env': 'CONNECTORS'
                    }
                ]
            })
            raise SystemExit(1)

        for conn_name in connectors_list:
            env_prefix = f'CONNECTOR_{conn_name.upper()}_'

            try:
                connectors[conn_name] = ConnectorSettings(_env_prefix=env_prefix, name=conn_name)

            except ValidationError as err:
                log.critical('Invalid connector settings', extra={
                    'connector': conn_name,
                    'errors': [
                        {
                            'error': error['msg'],
                            'env': f'{env_prefix}{".".join(str(x) for x in error["loc"]).upper()}'
                        } for error in err.errors()
                    ]
                })
                raise SystemExit(1)

        return settings, connectors


class RoutesLoader:
    @staticmethod
    def load(app, settings):
        from fetch_api.src.routes import internal, fleet_internal, fleet_api

        app.include_router(internal.router, prefix="/api")
        app.include_router(fleet_api.router, prefix="/car")
        app.include_router(fleet_internal.router)

        if 'grafana' in settings.connectors:
            from fetch_api.src.routes import grafana

            app.include_router(grafana.router, prefix="/grafana")
            log.info(f'Enabling Grafana routes at /grafana')

        if 'ml' in settings.connectors:
            from fetch_api.src.routes import ml

            app.include_router(ml.router, prefix="/ml")
            log.info(f'Enabling AI routes at /ml')

        log.info(f'Listening for incoming query requests on {settings.listen_host}:{settings.listen_port} and forwarding to upstream connectors')
