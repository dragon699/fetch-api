import os
from connectors.ml.src.providers import QueryResult, providers
from common.utils.system import read_file, render_template
from common.utils.helpers import TimeUtils
from common.telemetry.src.tracing.wrappers import traced
from common.telemetry.src.tracing.helpers import reword
from connectors.ml.settings import settings
from connectors.ml.src.telemetry.logging import log
from connectors.ml.src.query_processor import Processor



class Querier:
    def __init__(self, client) -> None:
        self.instructions_template_path = settings.instructions_template_path
        
        self.client = client
        self.instructions = {}

        self.load_instructions()


    def load_instructions(self) -> None:
        if not os.path.exists(self.instructions_template_path):
            log.critical(f'Instructions template file does not exist', extra={
                'instructions_file': self.instructions_template_path
            })
            raise SystemExit(1)

        try:
            instructions = read_file(self.instructions_template_path, type='yaml')

        except Exception as e:
            log.critical(f'{self.instructions_template_path} is not valid YAML file', extra={
                'instructions_file': self.instructions_template_path,
                'error': str(e)
            })
            raise SystemExit(1)

        self.instructions = instructions

    
    @traced('commit query')
    def commit(self, provider: str | None, prompt: str, model: str | None = None, instructions: str = '', instructions_template: str | None = None, span=None) -> dict:
        try:
            full_instructions = self.fetch(instructions, instructions_template)
            provider, model = self.client.resolve(provider, model)
            payload = self.render(prompt, model, full_instructions)
            response = self.send(provider, **payload)
            result = self.process(response)

            span.set_attributes(
                reword({
                    'querier.query.status': 'successful',
                    'querier.query.provider': provider,
                    'querier.query.prompt': payload['prompt'],
                    'querier.query.model': payload['model'],
                    'querier.query.instructions': payload['instructions'],
                    'querier.query.response': response.answer
                })
            )

            return result

        except Exception as err:
            span.set_attributes({
                'querier.query.status': 'failed',
                'querier.error.message': str(err),
                'querier.error.type': type(err).__name__
            })
            raise


    @traced('fetch instructions')
    def fetch(self, instructions: str, instructions_template: str | None, span=None) -> str:
        if instructions_template:
            span.set_attributes({
                'querier.instructions.template.path': self.instructions_template_path,
                'querier.instructions.template.key': instructions_template
            })

            if not instructions_template in self.instructions:
                log.warning(f'{instructions_template} is not a valid instructions template key')
                span.set_attributes({
                    'querier.instructions.template.key.exists': False
                })

            else:
                templated_instructions = render_template(
                    content=self.instructions[instructions_template],
                    vars={
                        'current_time': TimeUtils.time_now()
                    }
                )
                span.set_attributes({
                    'querier.instructions.template.key.exists': True,
                    'querier.query.instructions': templated_instructions
                })
                return templated_instructions

        if instructions:
            span.set_attributes({
                'querier.query.instructions': instructions
            })
            return instructions

        return ''


    @traced('render query payload')
    def render(self, prompt: str, model: str | None, instructions: str, span=None) -> dict:
        payload = {
            'prompt': prompt,
            'model': model or settings.default_model,
            'instructions': instructions or ''
        }

        span.set_attributes({
            'querier.query.prompt': prompt,
            'querier.query.model': payload['model'],
            'querier.query.instructions': instructions
        })

        return payload


    @traced('send query')
    def send(self, provider: str, prompt: str, model: str = None, instructions: str = '', span=None) -> QueryResult:
        return self.client.ask(provider, prompt, model, instructions)


    @traced('process query response')
    def process(self, response: QueryResult, span=None) -> dict:
        return Processor.process(response)


querier = Querier(providers)
