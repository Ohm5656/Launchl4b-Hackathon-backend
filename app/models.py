from pydantic import BaseModel

class EmailInput(BaseModel):
    from_field: str
    subject: str
    snippet: str
