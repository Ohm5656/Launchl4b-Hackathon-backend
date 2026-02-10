import os
from dotenv import load_dotenv
import requests
import rule_engine as re
import json

load_dotenv(dotenv_path=".env")

API_URL = "https://router.huggingface.co/v1/chat/completions"
HF_TOKEN = os.getenv("HF_TOKEN")

headers = {
    "Authorization": f"Bearer {HF_TOKEN}",
    "Content-Type": "application/json"
}

def buildPrompt(email):
    return f"""
You are classifying emails.

Decide if this email is about a paid subscription, billing, renewal, or charge.

Reply ONLY with:
YES or NO

Email:

From: {email['from']}
Subject: {email['subject']}
Body: {email['snippet']}
"""


def isSubscription(email):

    prompt = buildPrompt(email)

    payload = {
        "model": "meta-llama/Llama-3.1-8B-Instruct:novita",
        "messages": [
            {"role": "user", "content": prompt}
        ],
        "max_tokens": 5,
        "temperature": 0
    }

    r = requests.post(API_URL, headers=headers, json=payload)

    if r.status_code != 200:
        print("API Error:", r.text)
        return "NO"

    data = r.json()

    if "choices" not in data:
        print("Bad response:", data)
        return "NO"

    return data["choices"][0]["message"]["content"].strip()

email = {
    "from": "Google <no-reply@accounts.google.com>",
    "subject": "Subscription alert",
    "snippet": "You haven't paid your subscription..."
}

def returnJSON():
    result = isSubscription(email)
    print(json.dumps(re.rule_process(email)))
returnJSON()