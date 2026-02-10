import os
import json
from datetime import datetime, timezone
import requests
from dotenv import load_dotenv
import rule_engine as re

# =====================
# ENV
# =====================
load_dotenv(".env")

HF_API_URL = "https://router.huggingface.co/v1/chat/completions"
HF_TOKEN = os.getenv("HF_TOKEN")

FASTAPI_BASE = "http://localhost:8080"
FASTAPI_ENDPOINT = f"{FASTAPI_BASE}/api/ai/result"

headers = {
    "Authorization": f"Bearer {HF_TOKEN}",
    "Content-Type": "application/json"
}

# =====================
# LLM CLASSIFIER
# =====================
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
    payload = {
        "model": "meta-llama/Llama-3.1-8B-Instruct:novita",
        "messages": [{"role": "user", "content": buildPrompt(email)}],
        "max_tokens": 5,
        "temperature": 0
    }

    r = requests.post(HF_API_URL, headers=headers, json=payload)

    if r.status_code != 200:
        print("HF Error:", r.text)
        return False

    try:
        return r.json()["choices"][0]["message"]["content"].strip().upper() == "YES"
    except:
        return False

# =====================
# MAIN PIPELINE
# =====================
def analyze_and_send(emails):
    subscriptions = []

    for email in emails:
        if not isSubscription(email):
            continue

        # rule_engine ต้องคืน object แบบ subscription
        sub = re.rule_process(email)
        if sub:
            subscriptions.append(sub)

    payload = {
        "generated_at": datetime.now(timezone.utc).isoformat(),
        "subscriptions": subscriptions
    }

    print("📦 Payload to FastAPI:")
    print(json.dumps(payload, indent=2, ensure_ascii=False))

    # 🔥 ส่งไป FastAPI
    res = requests.post(
        FASTAPI_ENDPOINT,
        headers={"Content-Type": "application/json"},
        json=payload
    )

    if res.status_code >= 300:
        print("❌ Failed to send to FastAPI:", res.text)
    else:
        print("✅ Sent to FastAPI OK:", res.json())


# =====================
# TEST
# =====================
if __name__ == "__main__":
    emails = [
        {
            "id": "18c7f2b9a3e8c123",
            "from": "info@netflix.com",
            "subject": "Your Netflix receipt",
            "snippet": "You have been charged $15.49 for your monthly subscription"
        }
    ]

    analyze_and_send(emails)
