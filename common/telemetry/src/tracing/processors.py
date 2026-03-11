from urllib.parse import urlparse
from opentelemetry.sdk.trace import SpanProcessor, ReadableSpan
from opentelemetry.trace import Span
from opentelemetry.context import Context



class HttpSpanProcessor(SpanProcessor):
    def on_start(self, span: Span, parent_context: Context | None = None) -> None:
        if span.name in ['GET', 'POST', 'PUT', 'DELETE', 'PATCH', 'HEAD', 'OPTIONS']:
            url = span.attributes.get('http.url') or span.attributes.get('url.full')
            target = span.attributes.get('http.target') or span.attributes.get('url.path')

            if target:
                span.update_name(f'{span.name} {target}')

            elif url:
                span.update_name(f'{span.name} {urlparse(url).path}')


    def on_end(self, span: ReadableSpan) -> None:
        pass
    

    def shutdown(self) -> None:
        pass


    def force_flush(self, timeout_millis: int | None = None) -> bool:
        pass
