import atexit
import logging
from urllib.parse import urlparse
from fastapi import FastAPI
from opentelemetry import metrics, trace
from opentelemetry.exporter.otlp.proto.http.metric_exporter import OTLPMetricExporter
from opentelemetry.sdk.resources import Resource
from opentelemetry.sdk.metrics import MeterProvider
from opentelemetry.sdk.metrics.export import PeriodicExportingMetricReader
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from opentelemetry.instrumentation.fastapi import FastAPIInstrumentor
from opentelemetry.instrumentation.requests import RequestsInstrumentor
from opentelemetry.instrumentation.logging import LoggingInstrumentor
from common.telemetry.src.tracing.processors import HttpSpanProcessor
from common.telemetry.src.tracing.exporters import StatusSpanExporter



class Tracer:
    def __init__(self, otel_meta: dict, logger: logging.Logger) -> None:
        base_endpoint = self.normalize_otlp_endpoint(otel_meta['otlp_endpoint_http'])

        self.resource = Resource.create({
            'service.name': otel_meta['service_name'],
            'service.namespace': otel_meta['service_namespace'],
            'service.version': otel_meta['service_version']
        })

        self.span_exporter = StatusSpanExporter(
            endpoint=self.build_signal_endpoint(base_endpoint, 'traces')
        )
        self.metric_exporter = OTLPMetricExporter(
            endpoint=self.build_signal_endpoint(base_endpoint, 'metrics')
        )
        self.metric_reader = PeriodicExportingMetricReader(
            exporter=self.metric_exporter
        )

        self.tracer_provider = TracerProvider(resource=self.resource)
        self.tracer_provider.add_span_processor(
            BatchSpanProcessor(self.span_exporter)
        )
        self.tracer_provider.add_span_processor(
            HttpSpanProcessor()
        )

        self.meter_provider = MeterProvider(
            metric_readers=[self.metric_reader],
            resource=self.resource
        )

        trace.set_tracer_provider(self.tracer_provider)
        metrics.set_meter_provider(self.meter_provider)
        self.tracer = trace.get_tracer(otel_meta['service_name'])

        atexit.register(self.shutdown)
        logger.configure_otel()


    def instrument(self, app: FastAPI) -> None:
        FastAPIInstrumentor().instrument_app(
            app,
            tracer_provider=self.tracer_provider,
            meter_provider=self.meter_provider
        )
        RequestsInstrumentor().instrument(
            tracer_provider=self.tracer_provider,
            meter_provider=self.meter_provider
        )
        LoggingInstrumentor().instrument(set_logging_format=False)


    def get_tracer(self) -> trace.Tracer:
        return self.tracer


    def shutdown(self) -> None:
        self.tracer_provider.shutdown()
        self.meter_provider.shutdown()


    @staticmethod
    def normalize_otlp_endpoint(endpoint: str) -> str:
        if endpoint.startswith('http://') or endpoint.startswith('https://'):
            parsed = urlparse(endpoint)
            return f'{parsed.scheme}://{parsed.netloc}'

        parsed = urlparse(f'http://{endpoint}')
        if parsed.netloc:
            return f'http://{parsed.netloc}'

        return endpoint


    @staticmethod
    def build_signal_endpoint(base_endpoint: str, signal: str) -> str:
        clean = base_endpoint.rstrip('/')
        if clean.endswith(f'/v1/{signal}'):
            return clean
        if '/v1/' in clean:
            return clean
        return f'{clean}/v1/{signal}'
