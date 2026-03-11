from typing import Any
import os, json, yaml, jinja2



def read_file(path: str, type: str = 'json') -> Any:
    readers = {
        'yaml': lambda f: yaml.safe_load(f),
        'json': lambda f: json.load(f)
    }

    if not os.path.exists(path):
        return None

    with open(path, 'r') as file:
        file_type = path.split('.')[-1].lower()

        if file_type in ['yaml', 'yml'] or type in ['yaml', 'yml']:
            file_type = 'yaml'

        elif file_type in ['json'] or type in ['json']:
            file_type = 'json'

        with open(path, 'r') as f:
            try:
                return readers[file_type](f)

            except:
                return f.read()


def render_template(content: str, vars: dict | None = None) -> str:
    template = jinja2.Template(
        content,
        undefined=jinja2.StrictUndefined
    )

    return template.render(vars)
