from pydantic import BaseModel
from typing import Literal



class RequestAsk(BaseModel):
    provider: Literal['ollama', 'openclaw'] | None = None
    prompt: str
    model: str | None = None
    instructions: str | None = None
    instructions_template: str | None = None
