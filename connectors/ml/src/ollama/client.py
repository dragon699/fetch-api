import requests
from langchain_ollama import ChatOllama
from langchain_core.messages import BaseMessage
from connectors.ml.settings import settings
from common.telemetry.src.tracing.wrappers import traced
from common.telemetry.src.tracing.helpers import reword



class OllamaClient:
    def __init__(self) -> None:
        self.url = settings.url


    @staticmethod
    @traced('ping ollama')
    def ping(span=None) -> requests.Response:
        span.set_attributes({
            'ollama.operation': 'ping',
            'ollama.url': settings.url,
            'ollama.endpoint': settings.health_endpoint
        })

        response = requests.get(
            settings.health_endpoint,
            timeout=5
        )

        return response


    @traced('ask ollama')
    def ask_ollama(self, prompt: str, model: str, instructions: str = '', span=None) -> BaseMessage:
        span.set_attributes(
            reword({
                'ollama.operation': 'ask',
                'ollama.url': self.url,
                'ollama.model': model,
                'ollama.prompt': prompt,
                'ollama.instructions': instructions,
                'ollama.keep_alive': f'{settings.default_keep_alive_minutes}m',
                'ollama.temperature': settings.default_temperature
            })
        )

        model = ChatOllama(
            base_url=self.url,
            model=model,
            keep_alive=f'{settings.default_keep_alive_minutes}m',
            temperature=settings.default_temperature,
            num_ctx=10200,
            num_thread=8
        )

        response = model.invoke(
            [
                {
                    'role': 'system',
                    'content': instructions
                },
                {
                    'role': 'user',
                    'content': prompt
                }
            ]
        )

        return response
