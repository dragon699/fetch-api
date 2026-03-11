from pydantic import BaseModel



class RequestAsk(BaseModel):
    prompt: str
    model: str | None = None
    instructions: str | None = None
    instructions_template: str | None = None
