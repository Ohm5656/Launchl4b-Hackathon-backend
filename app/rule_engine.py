import re
from datetime import datetime
from typing import Optional


# =========================
# Helper extractors
# =========================

def extract_service_name(from_field: str) -> str:
    """
    info@netflix.com -> Netflix
    receipt@spotify.com -> Spotify
    """
    match = re.search(r'@([\w\-]+)\.', from_field.lower())
    if not match:
        return "Unknown"

    service = match.group(1)
    return service.capitalize()


def extract_amount(text: str) -> Optional[float]:
    """
    $15.49 / €9.99 / £12.00
    """
    match = re.search(r'([$€£])\s?(\d+(\.\d{2})?)', text)
    if not match:
        return None

    return float(match.group(2))


def extract_currency(text: str) -> Optional[str]:
    if "$" in text:
        return "USD"
    if "€" in text:
        return "EUR"
    if "£" in text:
        return "GBP"
    return None


def detect_billing_cycle(text: str) -> str:
    text = text.lower()
    if "year" in text or "annual" in text:
        return "yearly"
    return "monthly"


def detect_status(text: str) -> str:
    text = text.lower()

    if "payment failed" in text or "couldn't collect" in text:
        return "payment_failed"

    if "trial" in text:
        return "trial"

    if "receipt" in text or "invoice" in text or "charged" in text:
        return "active"

    return "unknown"


def detect_category(service_name: str) -> str:
    service = service_name.lower()

    if service in ["netflix", "spotify", "youtube"]:
        return "Streaming"
    if service in ["adobe", "figma", "notion"]:
        return "Productivity"
    if service in ["icloud", "google", "dropbox"]:
        return "Cloud"

    return "Other"


# =========================
# MAIN RULE ENGINE
# =========================

def rule_process(email: dict) -> dict:
    """
    email = {
      "id": "...",
      "from": "...",
      "subject": "...",
      "snippet": "..."
    }
    """

    combined_text = f"{email.get('subject', '')} {email.get('snippet', '')}"

    service_name = extract_service_name(email.get("from", ""))
    amount = extract_amount(combined_text)
    currency = extract_currency(combined_text)
    billing_cycle = detect_billing_cycle(combined_text)
    status = detect_status(combined_text)
    category = detect_category(service_name)

    return {
        # ===== core fields =====
        "service_name": service_name,
        "category": category,

        # วันที่: ถ้ายังหาไม่ได้ ให้ None (backend/AI pipeline จะจัดการต่อ)
        "subscribed_date": None,
        "next_billing_date": None,

        "billing_cycle": billing_cycle,
        "amount": amount,
        "currency": currency,
        "status": status,

        # ===== source tracking =====
        "source": {
            "email_id": email.get("id"),
            "from": email.get("from")
        }
    }
