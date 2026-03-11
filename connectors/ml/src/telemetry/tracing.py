from common.telemetry.tracer import Tracer
from connectors.ml.settings import settings
from connectors.ml.src.telemetry.logging import logger


instrumentor = Tracer(
    otel_meta={
        'service_name': settings.otel_service_name,
        'service_namespace': settings.otel_service_namespace,
        'service_version': settings.otel_service_version,
        'otlp_endpoint_grpc': settings.otlp_endpoint_grpc
    },
    logger=logger
)

tracer = instrumentor.get_tracer()
