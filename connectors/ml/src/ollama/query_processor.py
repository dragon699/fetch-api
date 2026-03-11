from langchain_core.messages import BaseMessage
from connectors.ml.settings import settings
from common.telemetry.src.tracing.wrappers import traced
from common.telemetry.src.tracing.helpers import reword



class Processor:
    @staticmethod
    @traced('process response')
    def process(response: BaseMessage, span=None) -> dict:
        result = []

        answer = response.content.rstrip('"').lstrip('"')
        if answer.startswith(' '):
            answer = answer[1:]

        result.append({
            'answer': answer,
            'duration_seconds': round((int(response.response_metadata['total_duration']) / 1_000_000_000), 1),
            'provider': response.response_metadata['model_provider'],
            'model': response.response_metadata['model_name'],
            'temperature': settings.default_temperature
        })

        span.set_attributes(
            reword({
                'processor.result.total_items': len(result),
                'processor.result.items': result
            })
        )

        return {
            'total_items': len(result),
            'items': result
        }
