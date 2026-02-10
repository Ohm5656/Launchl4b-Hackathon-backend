import re

def extract_service(from_field):
    match = re.search(r'@([\w\-]+)\.', from_field.lower())
    return match.group(1) if match else "unknown"

def extract_price(text):
    match = re.search(r'[$€£]\s?\d+(\.\d{2})?', text)
    return match.group() if match else None

def detect_status(text):
    text = text.lower()

    if "payment failed" in text or "couldn't collect" in text:
        return "payment_failed"

    if "renew" in text:
        return "renewal_notice"

    if "trial" in text:
        return "trial_ending"

    if "receipt" in text or "invoice" in text:
        return "receipt"

    return "unknown"

def calculate_confidence(service, price, status):
    score = 0
    if service != "unknown":
        score += 0.3
    if price:
        score += 0.3
    if status != "unknown":
        score += 0.4
    return score

def rule_process(email):
    combined = f"{email['subject']} {email['snippet']}"

    service = extract_service(email['from'])
    price = extract_price(combined)
    status = detect_status(combined)

    confidence = calculate_confidence(service, price, status)

    return {
        "service": service,
        "price": price,
        "renewal_date": None,
        "status": status,
        "confidence": confidence
    }
