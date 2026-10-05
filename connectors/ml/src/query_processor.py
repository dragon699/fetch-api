from dataclasses import asdict
from connectors.ml.src.providers import QueryResult


class Processor:
    @staticmethod
    def process(response: QueryResult) -> dict:
        item = asdict(response)
        if item['usage'] is None:
            del item['usage']
        return {'total_items': 1, 'items': [item]}
