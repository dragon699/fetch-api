import os, copy
from common.utils.system import read_file, render_template
from common.telemetry.src.tracing.wrappers import traced
from common.telemetry.src.tracing.helpers import reword
from connectors.grafana.settings import settings
from connectors.grafana.src.telemetry.logging import log
from connectors.grafana.src.api import GrafanaClient, grafana_client
from connectors.grafana.src.grafana.query_processor import Processor



class Querier:
    def __init__(self, client: GrafanaClient) -> None:
        self.templates_dir = settings.querier_templates_dir
        self.ds_uid_postgresql = settings.querier_ds_uid_postgresql
        self.ds_uid_prometheus = settings.querier_ds_uid_prometheus
        self.default_endpoint = settings.querier_default_endpoint

        self.client = client
        self.templates = {}

        self.set_templates_struct()
        self.load_templates()


    def set_templates_struct(self) -> None:
        self.templates_struct = {
            'prometheus': {
                'ds_uid': self.ds_uid_prometheus,
                'templates': ['payload.json', 'queries.yaml']
            },
            'postgresql': {
                'ds_uid': self.ds_uid_postgresql,
                'templates': ['payload.json', 'queries.yaml']
            }
        }


    def load_templates(self) -> None:
        for ds_dir_name in os.listdir(self.templates_dir):
            ds_dir_path = os.path.join(self.templates_dir, ds_dir_name)
            validated_files = []

            if not ds_dir_name in self.templates_struct:
                log.critical(f'{ds_dir_name} is not a valid ds templates directory', extra={
                    'valid_dirs': self.templates_struct.keys(),
                    'invalid_dir': ds_dir_path
                })
                raise SystemExit(1)

            if not ds_dir_name in self.templates:
                self.templates[ds_dir_name] = {}

            for ds_file_name in os.listdir(ds_dir_path):
                ds_template_name = ds_file_name.split('.')[0]
                ds_file_path = os.path.join(ds_dir_path, ds_file_name)

                if not ds_file_name in self.templates_struct[ds_dir_name]['templates']:
                    log.critical(f'{ds_file_name} is not a valid template file for {ds_dir_name}', extra={
                        'valid_files': self.templates_struct[ds_dir_name]['templates'],
                        'invalid_file': ds_file_path
                    })
                    raise SystemExit(1)

                if not ds_template_name in self.templates[ds_dir_name]:
                    self.templates[ds_dir_name][ds_template_name] = None

                try:
                    self.templates[ds_dir_name][ds_template_name] = read_file(ds_file_path)

                    if ds_template_name == 'payload':
                        self.templates[ds_dir_name][ds_template_name]['queries'][0]['datasource']['uid'] = self.templates_struct[ds_dir_name]['ds_uid']

                    validated_files.append(ds_file_name)

                except:
                    log.critical(f'{ds_file_name} is not a valid JSON file for {ds_dir_name}', extra={
                        'invalid_file': ds_file_path
                    })
                    raise SystemExit(1)
                
            if set(validated_files) != set(self.templates_struct[ds_dir_name]['templates']):
                log.critical(f'Missing template files for {ds_dir_name}', extra={
                    'validated_files': validated_files,
                    'missing_files': list(set(self.templates_struct[ds_dir_name]['templates']) - set(validated_files))
                })
                raise SystemExit(1)


    @traced('commit query')
    def commit(self, query_ds_type: str, query_id: str, query_params: dict | None = None, span=None) -> dict | None:
        query_params = query_params or {}

        span.set_attributes(
            reword({
                'querier.templates_dir': self.templates_dir,
                'querier.query.id': query_id,
                'querier.query.params': query_params,
                'querier.query.datasource.type': query_ds_type,
                'querier.query.datasource.uid': self.templates_struct[query_ds_type]['ds_uid']
            })
        )

        try:
            expression = self.fetch(query_ds_type, query_id, query_params)
            payload = self.render(query_ds_type, expression)
            response = self.send(payload)
            result = self.process(query_id, response.json())

            span.set_attributes({
                'querier.query.status': 'successful'
            })

            return result

        except Exception as err:
            span.set_attributes({
                'querier.query.status': 'failed',
                'querier.error.message': str(err),
                'querier.error.type': type(err).__name__
            })


    @traced('fetch query blueprint')
    def fetch(self, query_ds_type: str, query_id: str, query_params: dict | None = None, span=None) -> str | None:
        query_params = query_params or {}

        span.set_attributes({
            'querier.query.id': query_id,
            'querier.query.datasource.type': query_ds_type,
            'querier.query.datasource.uid': self.templates_struct[query_ds_type]['ds_uid']
        })

        if not query_ds_type in self.templates:
            span.set_attributes({
                'querier.error.message': f'No templates found for datasource type {query_ds_type}'
            })
            return None

        span.add_event('search started')

        for query in self.templates[query_ds_type]['queries']:
            if query['query']['id'] == query_id:
                expression = copy.deepcopy(query['query']['query'])

                span.add_event('search completed', attributes={
                    'querier.queries.template.path': f'{self.templates_dir}/{query_ds_type}/queries.yaml'
                })

                if len(query_params) > 0:
                    expression = render_template(
                        content=expression,
                        vars=query_params
                    )

                    span.set_attributes({
                        'querier.query.expression.templated': True,
                        'querier.query.expression': expression
                    })

                else:
                    span.set_attributes({
                        'querier.query.expression.templated': False,
                        'querier.query.expression': expression
                    })

                return expression

        span.set_attributes({
            'querier.error.message': f'No query found for id {query_id} in datasource type {query_ds_type}'
        })
        return None


    @traced('render query payload')
    def render(self, query_ds_type: str, expression: str,  span=None) -> dict:
        payload = copy.deepcopy(self.templates[query_ds_type]['payload'])
        expression_key = 'rawSql' if query_ds_type == 'postgresql' else 'expr'
        payload['queries'][0][expression_key] = expression

        span.set_attributes(
            reword({
                'querier.payload.template.path': f'{self.templates_dir}/{query_ds_type}/payload.json',
                'querier.query.expression.type': expression_key,
                'querier.query.payload': payload
            })
        )

        return payload


    @traced('send query')
    def send(self, query_payload: str, span=None) -> dict | None:
        try:
            response = self.client.post(
                endpoint = self.default_endpoint,
                data = query_payload
            )
            span.set_attributes(
                reword({
                    'querier.endpoint': self.default_endpoint,
                    'querier.query.payload': query_payload,
                    'querier.response.status_code': response.status_code,
                    'querier.response.content': response.text
                })
            )

            return response

        except Exception as err:
            span.set_attributes(
                reword({
                    'querier.error.message': f'Error occurred while sending query: {err}',
                    'querier.error.type': type(err).__name__,
                    'querier.endpoint': self.default_endpoint,
                    'querier.query.payload': query_payload
                })
            )
            return None


    @traced('process query response')
    def process(self, query_id: str, query_response: dict, span=None) -> dict | None:
        return Processor.process(query_id, query_response)


querier = Querier(grafana_client)
