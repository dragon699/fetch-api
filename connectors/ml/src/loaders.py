import json
from pydantic import ValidationError
from connectors.ml.src.telemetry.logging import log



class SettingsLoader:
    @staticmethod
    def load():
        from connectors.ml.settings import Settings

        try:
            settings = Settings()

        except ValidationError as err:
            print(json.dumps([
                {
                    'error': error['msg'],
                    'env': f'{error["loc"][0].upper()}'
                } for error in err.errors()
            ], indent=2))
            raise SystemExit(1)
        
        return settings


class RoutesLoader:
    @staticmethod
    def load(app, settings):
        from connectors.ml.src.routes import (internal, ask)

        app.include_router(internal.router, prefix="/api")
        app.include_router(ask.router, prefix="/ask")
        log.info(
            f'Listening for incoming query requests on {settings.listen_host}:{settings.listen_port} and forwarding to Ollama at {settings.url}'
        )
