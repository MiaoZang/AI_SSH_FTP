---
name: ai-ssh-ftp-proxy
description: "AI Agent Skill for executing SSH commands and file operations on remote servers via a proxy service. Supports HTTP API and WebSocket for interactive sessions."
version: "1.4.0"
---

# AI SSH/FTP Proxy Skill

Enable AI agents to securely access remote servers via SSH and HTTP file transfer.

## Overview

This skill provides a proxy service that AI agents can call to:
- Execute shell commands on remote servers (SSH)
- **Upload files/folders with auto-extract** (HTTP Multipart) ✨
- **File operations**: mkdir, rename, copy, info, batch delete ✨
- Open interactive shell sessions (WebSocket)

All inputs and outputs are Base64 encoded for safe transmission.

---

## ⚠️ First-Time Deployment (首次部署)

> [!IMPORTANT]
> If the target server does NOT have this proxy service installed yet, follow the steps below.
> If already installed, skip to [API Endpoints](#api-endpoints).

### Step 1: Connect to Server via SSH

Use the **SSH 直连方式** skill (`.agent/skills/JN___Open SSH连接`) to connect:

```powershell
C:\Windows\System32\OpenSSH\ssh.exe -o StrictHostKeyChecking=no -p {SSH_PORT} {USER}@{SERVER_IP}
```

### Step 2: Run the Install Script

After SSH login, execute:

```bash
# Download and run installer (one-liner)
curl -fsSL https://raw.githubusercontent.com/MiaoZang/AI_SSH_FTP/main/scripts/install.sh -o /tmp/install.sh && \
chmod +x /tmp/install.sh && \
/tmp/install.sh --ssh-port {SSH_PORT} --ssh-user {USER} --ssh-pass {PASSWORD} --auto-start --systemd
```

> [!TIP]
> - `--ssh-host` defaults to `127.0.0.1` (proxy connects to localhost SSH)
> - `--ssh-port` should match the server's actual SSH port
> - `--ssh-user` and `--ssh-pass` = the same credentials you used to SSH in
> - `--auto-start` starts the service immediately after install
> - `--systemd` registers as a systemd service for auto-restart

### Step 3: Verify

```bash
curl http://127.0.0.1:48891/api/health
# Expected: {"status":"ok"}
```

### Full Install Parameters

| Parameter | Default | Description |
|-----------|---------|-------------|
| `--ssh-pass` | (required) | SSH password |
| `--ssh-host` | `127.0.0.1` | SSH host |
| `--ssh-port` | `22` | SSH port |
| `--ssh-user` | `root` | SSH username |
| `--http-port` | `48891` | HTTP API port |
| `--ws-port` | `48892` | WebSocket port |
| `--install-dir` | `/opt/ssh-ftp-proxy` | Install directory |
| `--auto-start` | false | Start after install |
| `--systemd` | false | Create systemd service |
| `--force` | false | Force reinstall |

---

## API Endpoints

### Health Check

```bash
curl http://SERVER:48891/api/health
```

### SSH Command Execution

**Method 1: GET (recommended for PowerShell / simple calls)**

```bash
# cmd = base64 encoded command
curl "http://SERVER:48891/api/ssh/exec?cmd=BASE64_COMMAND"
```

**Method 2: POST (JSON body)**

```bash
curl -X POST http://SERVER:48891/api/ssh/exec \
  -H "Content-Type: application/json" \
  -d '{"command": "BASE64_ENCODED_COMMAND"}'
```

**Method 3: PowerShell Helper Script (recommended for Windows)**

```powershell
.\ssh_exec.ps1 -Command "ls -la /" -Server "http://SERVER:48891"
.\ssh_exec.ps1 -Command "pm2 restart all" -Server "http://SERVER:48891"
```

Response:
```json
{"stdout": "BASE64_OUTPUT", "stderr": "BASE64_ERRORS", "exit_code": 0}
```

---

### Windows / PowerShell Usage Guide

> [!TIP]
> PowerShell has JSON escaping issues with `curl.exe`. Use one of these methods instead:

**Option A: GET API (simplest)**
```powershell
$cmd = [Convert]::ToBase64String([Text.Encoding]::UTF8.GetBytes("ls -la /"))
$resp = Invoke-RestMethod -Uri "http://SERVER:48891/api/ssh/exec?cmd=$cmd" -Method GET
$output = [Text.Encoding]::UTF8.GetString([Convert]::FromBase64String($resp.stdout))
Write-Host $output
```

**Option B: ssh_exec.ps1 helper script**
```powershell
# Download
Invoke-WebRequest -Uri "https://raw.githubusercontent.com/MiaoZang/AI_SSH_FTP/main/scripts/ssh_exec.ps1" -OutFile ssh_exec.ps1

# Use
.\ssh_exec.ps1 -Command "ls -la /" -Server "http://SERVER:48891"
```

**Option C: POST with temp file (for complex commands)**
```powershell
$cmd = [Convert]::ToBase64String([Text.Encoding]::UTF8.GetBytes("ls -la /"))
@{command=$cmd} | ConvertTo-Json | Set-Content cmd.json -Encoding UTF8
curl.exe -s -X POST http://SERVER:48891/api/ssh/exec -H "Content-Type: application/json" -d "@cmd.json"
Remove-Item cmd.json
```

---


### File Upload API (HTTP Multipart)

#### Upload & Auto-Extract (recommended for deployment)

```bash
curl -X POST http://SERVER:48891/api/file/upload \
  -F "file=@archive.tar.gz" \
  -F "path=BASE64_DEST_PATH" \
  -F "extract=true"
```

> 💡 **Tip**: path ending with `/` auto-appends the filename

#### List Directory

```bash
curl -X POST http://SERVER:48891/api/file/list \
  -H "Content-Type: application/json" \
  -d '{"path": "BASE64_PATH"}'
```

#### Download / Delete File

```bash
# Download
curl -X POST http://SERVER:48891/api/file/download \
  -H "Content-Type: application/json" \
  -d '{"path": "BASE64_PATH"}'

# Delete
curl -X POST http://SERVER:48891/api/file/delete \
  -H "Content-Type: application/json" \
  -d '{"path": "BASE64_PATH"}'
```

---

### File Operations

#### Create Directory

```bash
curl -X POST http://SERVER:48891/api/file/mkdir \
  -H "Content-Type: application/json" \
  -d '{"path": "BASE64_PATH"}'
```

#### Rename / Move

```bash
curl -X POST http://SERVER:48891/api/file/rename \
  -H "Content-Type: application/json" \
  -d '{"src": "BASE64_SRC", "dst": "BASE64_DST"}'
```

#### Copy

```bash
curl -X POST http://SERVER:48891/api/file/copy \
  -H "Content-Type: application/json" \
  -d '{"src": "BASE64_SRC", "dst": "BASE64_DST"}'
```

#### File Info

```bash
curl -X POST http://SERVER:48891/api/file/info \
  -H "Content-Type: application/json" \
  -d '{"path": "BASE64_PATH"}'
```

#### Batch Delete

```bash
curl -X POST http://SERVER:48891/api/file/batch/delete \
  -H "Content-Type: application/json" \
  -d '{"paths": ["BASE64_PATH1", "BASE64_PATH2"]}'
```

---

### WebSocket Interactive SSH

Connect to `ws://SERVER:48892/ws/ssh`

```json
// Client → Server
{"type": "input", "payload": "BASE64_INPUT"}
// Server → Client
{"type": "output", "payload": "BASE64_OUTPUT"}
```

---

## Practical Examples

### Example 1: Deploy Project via Script

```bash
# Linux/Mac
./deploy.sh ./dist /www/wwwroot/app/ http://SERVER:48891

# Windows PowerShell
.\deploy.ps1 -LocalDir .\dist -RemotePath /www/wwwroot/app/ -ServerUrl http://SERVER:48891
```

### Example 2: Deploy via API

```bash
# 1. Compress
tar -czvf dist.tar.gz ./dist

# 2. Encode path
echo -n "/www/wwwroot/app/" | base64

# 3. Upload & extract
curl -X POST http://SERVER:48891/api/file/upload \
  -F "file=@dist.tar.gz" \
  -F "path=L3d3dy93d3dyb290L2FwcC8=" \
  -F "extract=true"
```

### Example 3: PowerShell Workflow

```powershell
$path = [Convert]::ToBase64String([Text.Encoding]::UTF8.GetBytes("/www/wwwroot/app/"))
curl.exe -X POST http://SERVER:48891/api/file/upload `
  -F "file=@dist.tar.gz" `
  -F "path=$path" `
  -F "extract=true"
```

---

## Configuration

Edit `config/config.yaml`:
```yaml
server:
  http_port: 48891
  ws_port: 48892
  bind_ip: "0.0.0.0"

ssh_server:
  host: "127.0.0.1"
  port: 22
  user: "root"
  password: "password"
```

---

## Scripts

| Script | Purpose | Mode |
|--------|---------|------|
| `scripts/install.sh` | Install & configure service | CLI (AI-friendly) |
| `scripts/manage.sh` | Service management | Interactive menu |
| `scripts/deploy.sh` | Deploy local project to server | CLI |
| `scripts/deploy.ps1` | Deploy local project (Windows) | CLI |

---

## Version History

### v1.4.0 (2026-02-19)
- ✨ **AI Install Script** - Non-interactive `install.sh` with CLI parameters
- ✨ **Systemd support** - Auto-restart on failure
- 📖 Updated SKILL.md with first-time deployment guide

### v1.3.0 (2026-02-10)
- ✨ **File operations** - mkdir, rename, copy, info, batch delete
- ✅ All file APIs tested and verified

### v1.2.0 (2026-02-09)
- ✨ **HTTP File Upload API** - Multipart upload, no FTP required
- ✨ **Auto-extract support** - tar.gz, zip, tar

### v1.1.0 (2026-02-08)
- ✅ Fixed SSH connection race condition
- ✅ Fixed WebSocket goroutine leak

### v1.0.0 (2026-01-18)
- Initial release

---

## Repository

**GitHub**: https://github.com/MiaoZang/AI_SSH_FTP

## Security Notes

- Credentials stored server-side, not passed per-request
- All data Base64 encoded to prevent injection
- Deploy on private network or use firewall rules
- Consider SSH key authentication over passwords
