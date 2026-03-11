import atexit
import logging
from fastapi import FastAPI
from opentelemetry import trace
from opentelemetry.sdk.resources import Resource
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from opentelemetry.instrumentation.fastapi import FastAPIInstrumentor
from opentelemetry.instrumentation.requests import RequestsInstrumentor
from opentelemetry.instrumentation.logging import LoggingInstrumentor
from common.telemetry.src.tracing.processors import HttpSpanProcessor
from common.telemetry.src.tracing.exporters import StatusSpanExporter



class Tracer:
    def __init__(self, otel_meta: dict, logger: logging.Logger) -> None:
        self.resource = Resource.create({
            'service.name': otel_meta['service_name'],
            'service.namespace': otel_meta['service_namespace'],
            'service.version': otel_meta['service_version']
        })
        self.span_exporter = StatusSpanExporter(
            endpoint = otel_meta['otlp_endpoint_grpc'],
            insecure = True
        )

        self.provider = TracerProvider(resource=self.resource)
        self.provider.add_span_processor(
            BatchSpanProcessor(self.span_exporter)
        )
        self.provider.add_span_processor(
            HttpSpanProcessor()
        )

        trace.set_tracer_provider(self.provider)
        self.tracer = trace.get_tracer(otel_meta['service_name'])

        atexit.register(lambda: self.provider.shutdown())
        logger.configure_otel()


    def instrument(self, app: FastAPI) -> None:
        FastAPIInstrumentor().instrument_app(app)
        RequestsInstrumentor().instrument()
        LoggingInstrumentor().instrument(set_logging_format=False)


    def get_tracer(self) -> trace.Tracer:
        return self.tracer
