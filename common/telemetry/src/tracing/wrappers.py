from typing import Any, Callable
from functools import wraps
from os import getenv
from opentelemetry import trace


tracer = trace.get_tracer(
    getenv("OTEL_SERVICE_NAME", __name__)
)


def traced(
    span_name: str | None = None,
    attributes: dict | None = None,
    inject_span: bool = True
) -> Callable[[Callable[..., Any]], Callable[..., Any]]:
    def decorator(func: Callable[..., Any]) -> Callable[..., Any]:
        @wraps(func)
        def wrapper(*args: Any, **kwargs: Any) -> Any:
            if span_name:
                name = span_name

            else:
                if args and hasattr(args[0].__class__, func.__name__):
                    name = f"{args[0].__class__.__name__}.{func.__name__}"

                elif hasattr(func, '__qualname__') and '.' in func.__qualname__:
                    name = func.__qualname__

                else:
                    name = f"{func.__module__}.{func.__name__}"

            with tracer.start_as_current_span(name) as span:
                span.set_attribute("func.module", func.__module__)

                if attributes:
                    for key, value in attributes.items():
                        span.set_attribute(key, value)

                try:
                    if inject_span:
                        kwargs['span'] = span

                    result = func(*args, **kwargs)
                    span.set_status(
                        trace.Status(trace.StatusCode.OK)
                    )

                    return result

                except Exception as e:
                    span.set_status(
                        trace.Status(trace.StatusCode.ERROR)
                    )
                    span.record_exception(e)

                    raise

        return wrapper
    return decorator
