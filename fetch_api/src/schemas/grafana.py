from pydantic import BaseModel



class GrafanaBody(BaseModel):
    ai: bool = False
