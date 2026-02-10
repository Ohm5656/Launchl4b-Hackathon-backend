from fastapi import FastAPI, HTTPException
from fastapi.middleware.cors import CORSMiddleware
from pathlib import Path
from typing import List, Optional, Dict, Any
from pydantic import BaseModel
import json
from datetime import datetime

# ===== ของเดิมคุณ =====
from .models import EmailInput
from .rule_engine import rule_process

app = FastAPI(title="SubTrack AI Backend (FastAPI)")

# ===== CORS (ให้ Frontend เรียกได้) =====
app.add_middleware(
    CORSMiddleware,
    allow_origins=[
        "http://localhost:5173",  # Vite
        "http://localhost:3000"
    ],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# ===== Path สำหรับเก็บผลลัพธ์ =====
OUTPUT_DIR = Path("output")
OUTPUT_DIR.mkdir(exist_ok=True)
SUBSCRIPTION_FILE = OUTPUT_DIR / "subscriptions.json"

# =====================================================
# 1️⃣ MODEL สำหรับ subscription (ตรงกับ JSON ที่คุณออกแบบ)
# =====================================================

class Source(BaseModel):
    email_id: Optional[str] = None
    from_email: Optional[str] = None

class Subscription(BaseModel):
    service_name: str
    category: Optional[str] = None
    subscribed_date: Optional[str] = None
    next_billing_date: Optional[str] = None
    billing_cycle: Optional[str] = None
    amount: Optional[float] = None
    currency: Optional[str] = None
    status: Optional[str] = None
    source: Optional[Dict[str, Any]] = None

class AIResult(BaseModel):
    generated_at: Optional[str] = None
    subscriptions: List[Subscription]

# =====================================================
# 2️⃣ Endpoint เดิม: วิเคราะห์ email ด้วย rule_engine
# =====================================================
@app.post("/process")
def process(email: EmailInput):
    """
    รับ email → วิเคราะห์ด้วย AI/rule_engine
    """
    result = rule_process(email)
    return result

# =====================================================
# 3️⃣ Endpoint รับผล AI (JSON) แล้วเก็บไฟล์
# =====================================================
@app.post("/api/ai/result")
def receive_ai_result(payload: AIResult):
    data = {
        "generated_at": payload.generated_at or datetime.utcnow().isoformat(),
        "subscriptions": payload.subscriptions
    }

    SUBSCRIPTION_FILE.write_text(
        json.dumps(data, indent=2, ensure_ascii=False),
        encoding="utf-8"
    )

    return {
        "status": "ok",
        "saved_to": str(SUBSCRIPTION_FILE),
        "count": len(payload.subscriptions)
    }

# =====================================================
# 4️⃣ Endpoint ให้ Frontend ดึงไปแสดง
# =====================================================
@app.get("/api/subscriptions")
def get_subscriptions():
    """
    Frontend เรียก endpoint นี้เพื่อเอาไปแสดงบน Dashboard
    """
    if not SUBSCRIPTION_FILE.exists():
        raise HTTPException(status_code=404, detail="No subscription data yet")

    data = json.loads(SUBSCRIPTION_FILE.read_text(encoding="utf-8"))

    # ส่งเฉพาะ list → frontend ใช้ง่าย
    return data.get("subscriptions", [])

# =====================================================
# 5️⃣ Health check (แนะนำ)
# =====================================================
@app.get("/api/health")
def health():
    return {"ok": True}
