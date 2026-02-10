from fastapi import FastAPI
from .models import EmailInput
from .rule_engine import rule_process

app = FastAPI()

@app.post("/process")
def process(email: EmailInput):
    return rule_process(email)
