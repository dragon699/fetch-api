import logging


class Filters:
    class Drop(logging.Filter):
        def filter(self, record: logging.LogRecord) -> bool:
            try:
                if (record.name == 'uvicorn.access') and ('/api/health' in record.getMessage()):
                    return False

            except Exception:
                return True

            return True


    class Clear(logging.Filter):
        def filter(self, record: logging.LogRecord) -> bool:
            deleted_fields = ['otelTraceSampled', 'otelServiceName']
            checked_fields = ['trace_id', 'span_id']

            for field in deleted_fields:
                if hasattr(record, field):
                    delattr(record, field)

            for field in checked_fields:
                if hasattr(record, field) and getattr(record, field) == '0':
                    delattr(record, field)

            return True


    class Rename(logging.Filter):
        def filter(self, record: logging.LogRecord) -> bool:
            renamed_fields = {
                'levelname': 'level',
                'name': 'component',
                'message': 'msg',
                'asctime': 'time',
                'otelTraceID': 'trace_id',
                'otelSpanID': 'span_id'
            }

            preserved = {'name', 'levelname'}

            for current_field, renamed_field in renamed_fields.items():
                if hasattr(record, current_field):
                    setattr(record, renamed_field, getattr(record, current_field))
                    if not current_field in preserved:
                        delattr(record, current_field)

            return True
