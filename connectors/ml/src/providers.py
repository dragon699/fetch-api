from dataclasses import dataclass
from time import perf_counter

import requests
from connectors.ml.settings import settings
from connectors.ml.src.ollama.client import OllamaClient


@dataclass
class QueryResult:
    answer: str
    provider: str
    model: str
    duration_seconds: float
    temperature: float
    usage: dict | None = None


class OpenClawClient:
    def ask(self, prompt: str, model: str, instructions: str = '') -> dict:
        if not settings.openclaw_url or not settings.openclaw_token:
            raise ValueError('OpenClaw requires OPENCLAW_URL and OPENCLAW_TOKEN')
        response = requests.post(
            settings.openclaw_url.rstrip('/') + '/v1/chat/completions',
            headers={'Authorization': f'Bearer {settings.openclaw_token}'},
            json={
                'model': model,
                'temperature': settings.default_temperature,
                'messages': [
                    {'role': 'system', 'content': instructions},
                    {'role': 'user', 'content': prompt}
                ]
            },
            timeout=(5, settings.openclaw_timeout_seconds)
        )
        response.raise_for_status()
        data = response.json()
        content = data['choices'][0]['message']['content']
        if not isinstance(content, str) or not content.strip():
            raise ValueError('OpenClaw returned no text answer')
        return data


class ProviderRegistry:
    def __init__(self):
        self.clients = {'ollama': OllamaClient(), 'openclaw': OpenClawClient()}

    def resolve(self, provider=None, model=None):
        provider = provider or settings.default_provider
        if provider not in self.clients:
            raise ValueError(f'Unsupported provider: {provider}')
        default_model = (
            settings.ollama_default_model or settings.default_model
            if provider == 'ollama' else settings.openclaw_default_model
        )
        model = model or default_model
        if not model:
            raise ValueError(f'No default model configured for {provider}')
        return provider, model

    def configuration(self):
        return {
            'default_provider': settings.default_provider,
            'providers': {
                provider: {'default_model': (settings.ollama_default_model or settings.default_model
                                            if provider == 'ollama' else settings.openclaw_default_model)}
                for provider in self.clients
            }
        }

    def ask(self, provider, prompt, model, instructions=''):
        provider, model = self.resolve(provider, model)
        if provider == 'ollama' and not settings.url:
            raise ValueError('Ollama requires OLLAMA_URL or URL')
        start = perf_counter()
        response = self.clients[provider].ask(prompt, model, instructions)
        if provider == 'ollama':
            answer = response.content
            usage = response.usage_metadata
        else:
            answer = response['choices'][0]['message']['content']
            usage = response.get('usage')
        if not isinstance(answer, str) or not answer.strip():
            raise ValueError(f'{provider} returned no text answer')
        return QueryResult(answer, provider, model, round(perf_counter() - start, 3),
                           settings.default_temperature, usage)

    def ping(self, provider):
        if provider == 'ollama':
            if not settings.url:
                return False
            response = OllamaClient.ping()
            return response.status_code == 200 and response.text == 'Ollama is running'
        if not settings.openclaw_url or not settings.openclaw_token:
            return False
        response = requests.get(
            settings.openclaw_url.rstrip('/') + '/health', timeout=5,
            headers={'Authorization': f'Bearer {settings.openclaw_token}'}
        )
        return response.status_code == 200 and response.json().get('ok') is True


providers = ProviderRegistry()
