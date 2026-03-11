import logging
from common.telemetry.src.logging.formatters import Formatters
from common.telemetry.src.logging.filters import Filters



class Logger:
    def __init__(self, name: str) -> None:
        self.name = name
        self.log_level = None
        self.log_format = None

        self.logger = logging.getLogger(name)
        self.style = {'datefmt': '%Y-%m-%dT%H:%M:%S'}

        self.handler = logging.StreamHandler()
        self.handler.addFilter(Filters.Clear())
        self.handler.addFilter(Filters.Rename())

        self.formatter = None
        self.logger.addHandler(self.handler)


    def create_formatter(self, log_format: str) -> logging.Formatter:
        if log_format.lower() == 'logfmt':
            return Formatters.Logfmt(
                datefmt=self.style['datefmt']
            )

        else:
            return Formatters.Json(
                fmt='%(asctime)s %(level)s %(component)s %(msg)s',
                datefmt=self.style['datefmt'],
                rename_fields={'asctime': 'time'}
            )


    def update_settings(self, log_level: str, log_format: str) -> None:
        self.log_level = log_level
        self.log_format = log_format

        self.logger.setLevel(log_level.upper())
        self.formatter = self.create_formatter(log_format)
        self.handler.setFormatter(self.formatter)
        self.logger.addHandler(self.handler)


    def get(self) -> logging.Logger:
        return self.logger
    

    def get_uvicorn_config(self) -> dict:
        if self.log_format.lower() == 'json':
            formatter_config = {
                '()': 'common.telemetry.src.logging.formatters.Formatters.Json',
                'fmt': '%(asctime)s %(level)s %(component)s %(msg)s',
                'datefmt': '%Y-%m-%dT%H:%M:%S',
                'rename_fields': {
                    'asctime': 'time'
                }
            }
        else:
            formatter_config = {
                '()': 'common.telemetry.src.logging.formatters.Formatters.Logfmt',
                'datefmt': '%Y-%m-%dT%H:%M:%S'
            }

        return {
            'version': 1,
            'disable_existing_loggers': False,
            'filters': {
                'Drop': {
                    '()': 'common.telemetry.src.logging.filters.Filters.Drop'
                },
                'Clear': {
                    '()': 'common.telemetry.src.logging.filters.Filters.Clear'
                },
                'Rename': {
                    '()': 'common.telemetry.src.logging.filters.Filters.Rename'
                }
            },
            'formatters': {
                'default': {
                    **formatter_config
                }
            },
            'handlers': {
                'default': {
                    'formatter': 'default',
                    'class': 'logging.StreamHandler',
                    'filters': ['Drop', 'Clear', 'Rename']
                }
            },
            'loggers': {
                'uvicorn': {
                    'handlers': ['default'],
                    'level': 'WARNING'
                },
                'uvicorn.access': {
                    'handlers': ['default'],
                    'level': 'INFO',
                    'propagate': False
                }
            }
        }


    def configure_otel(self) -> None:
        loggers = [
            'opentelemetry.sdk.trace',
            'opentelemetry.exporter.otlp',
            'opentelemetry.instrumentation',
        ]

        for logger_name in loggers:
            otel_logger = logging.getLogger(logger_name)
            otel_logger.handlers.clear()

            otel_logger.addHandler(self.handler)
            otel_logger.setLevel(self.log_level.upper())

            otel_logger.propagate = False
