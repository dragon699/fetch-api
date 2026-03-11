import logging
from pythonjsonlogger.jsonlogger import JsonFormatter


class Formatters:
    class Logfmt(logging.Formatter):
        def format(self, record: logging.LogRecord) -> str:
            attr = [
                f'time="{self.formatTime(record, self.datefmt)}"',
                f'level={record.level}',
                f'component={record.component}',
                f'msg="{record.getMessage()}"'
            ]

            if hasattr(record, '__dict__'):
                for key, value in record.__dict__.items():
                    if key in ['component', 'msg', 'args', 'created', 'filename', 'funcName', 
                              'level', 'levelname', 'levelno', 'lineno', 'module', 'msecs', 
                              'message', 'name', 'pathname', 'process', 'processName', 'relativeCreated',
                              'thread', 'threadName', 'exc_info', 'exc_text', 'stack_info',
                              'asctime', 'taskName']:
                        continue

                    if isinstance(value, str):
                        attr.append(f'{key}="{value}"')

                    elif value is None:
                        attr.append(f'{key}=null')

                    elif isinstance(value, bool):
                        attr.append(f'{key}={str(value).lower()}')

                    else:
                        attr.append(f'{key}={value}')

            return ' '.join(attr)


    class Json(JsonFormatter):
        def add_fields(self, log_record: dict, record: logging.LogRecord, message_dict: dict) -> None:
            super().add_fields(log_record, record, message_dict)

            log_record.pop('name', None)
            log_record.pop('levelname', None)

            if 'msg' in log_record:
                log_record['msg'] = record.getMessage()
