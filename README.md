# Gmail Inbox Fetcher Web Pro 🚀

A premium web application to automate fetching your Gmail inbox messages. No complex terminal commands for users—just click, authorize, and see your messages.

## 🌟 Key Features
- **One-Click Authorization:** Clean web-based OAuth2 flow.
- **Premium UI:** Modern, dark-mode design with glassmorphism aesthetics.
- **Automated Messaging Fetching:** Retrieves the last 20 messages (From, Subject, Snippet).
- **Auto-JSON Export:** Saves the results to `inbox.json` automatically.

## 🛠️ Setup Instructions

### 1. Google Cloud Console Configuration
To make this app work, you need to configure your Google Cloud Project:
1. Go to [Google Cloud Console](https://console.cloud.google.com/).
2. **Enable Gmail API** for your project.
3. **OAuth Consent Screen:**
   - Set User Type to **External**.
   - Add your email to **Test Users** (Critical!).
4. **Create Credentials:**
   - Type: **OAuth client ID**.
   - Application Type: **Web application**.
   - **Authorized redirect URIs:** Add `http://localhost:8080/callback`.
5. **Download JSON:** Save the credentials file as `credentials.json` in the root folder of this project.

### 2. Run the Application
```bash
# Install dependencies
go mod tidy

# Run the server
go run main.go
```

### 3. Usage
1. Open your browser to `http://localhost:8080`.
2. Click **"Continue with Google"**.
3. Authorize the app.
4. Your inbox messages will be displayed on the screen and saved to `inbox.json`.

## 📁 Project Structure
- `cmd/server/main.go`: จุดเริ่มต้นของโปรแกรม (Entry Point)
- `internal/auth/`: จัดการการตั้งค่า OAuth2
- `internal/gmail/`: จัดการการดึงข้อมูลจาก Gmail API (Concurrency)
- `internal/handlers/`: จัดการ HTTP Routes และ Business Logic
- `templates/`: เก็บหน้า UI (HTML/CSS)
- `output/`: เก็บไฟล์ผลลัพธ์ (`inbox.json`)

## 🛠️ Setup & Run
```bash
# รันโปรแกรม
go run cmd/server/main.go
```
