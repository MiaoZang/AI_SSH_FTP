---
name: ai-ssh-ftp-proxy
description: "AI Agent Skill for executing SSH commands and file operations on remote servers via a proxy service. Supports HTTP API and WebSocket for interactive sessions."
version: "1.2.0"
---

# AI SSH/FTP Proxy Skill

Enable AI agents to securely access remote servers via SSH and HTTP file transfer.

## Overview

This skill provides a proxy service that AI agents can call to:
- Execute shell commands on remote servers (SSH)
- **Upload files/folders with auto-extract** (HTTP Multipart) ✨
- Open interactive shell sessions (WebSocket)

All inputs and outputs are Base64 encoded for safe transmission.

## Quick Start

### One-Line Installation

```bash
curl -fsSL https://raw.githubusercontent.com/MiaoZang/AI_SSH_FTP/main/scripts/manage.sh | bash
```

### Management Script Features

- 🌐 **Bilingual** - English and Chinese support
- 📦 **Auto-download** - Downloads binary from GitHub if missing
- 🔧 **Interactive config** - Creates config.yaml via wizard

### One-Click Deploy Script ✨ NEW

**AI 下载脚本 / AI Download Scripts:**
```bash
# Linux/Mac - 下载 deploy.sh
wget https://raw.githubusercontent.com/MiaoZang/AI_SSH_FTP/main/scripts/deploy.sh
chmod +x deploy.sh

# Windows PowerShell - 下载 deploy.ps1
Invoke-WebRequest -Uri "https://raw.githubusercontent.com/MiaoZang/AI_SSH_FTP/main/scripts/deploy.ps1" -OutFile deploy.ps1
```

**使用方法 / Usage:**

Linux/Mac:
```bash
./deploy.sh <本地目录> <远程路径> <服务器地址>
./deploy.sh ./dist /www/wwwroot/app/ http://SERVER:48891
```

Windows PowerShell:
```powershell
.\deploy.ps1 -LocalDir <本地目录> -RemotePath <远程路径> -ServerUrl <服务器地址>
.\deploy.ps1 -LocalDir .\dist -RemotePath /www/wwwroot/app/ -ServerUrl http://SERVER:48891
```

**功能 / Features:**
- 📦 自动压缩本地目录 / Auto-compress local directory
- 🚀 上传到远程服务器 / Upload to remote server  
- 📂 自动解压到目标路径 / Auto-extract to destination
- ✅ 验证部署结果 / Verify deployment


---

## API Endpoints

### SSH Command Execution

```bash
curl -X POST http://SERVER:48891/api/ssh/exec \
  -H "Content-Type: application/json" \
  -d '{"command": "BASE64_ENCODED_COMMAND"}'
```

Response:
```json
{"stdout": "BASE64_OUTPUT", "stderr": "BASE64_ERRORS", "exit_code": 0}
```

---

### File Upload API (HTTP Multipart) ✨ NEW

#### Upload File

```bash
# path = base64 encoded destination path
curl -X POST http://SERVER:48891/api/file/upload \
  -F "file=@local_file.tar.gz" \
  -F "path=BASE64_DEST_PATH"
```

#### Upload & Auto-Extract (推荐用于文件夹部署)

```bash
curl -X POST http://SERVER:48891/api/file/upload \
  -F "file=@archive.tar.gz" \
  -F "path=BASE64_DEST_PATH" \
  -F "extract=true"
```

> 💡 **Tip**: 目标路径以 `/` 结尾会自动追加文件名

Response:
```json
{"success": true, "path": "/www/wwwroot/app/archive.tar.gz", "size": 493518}
```

#### List Directory

```bash
curl -X POST http://SERVER:48891/api/file/list \
  -H "Content-Type: application/json" \
  -d '{"path": "BASE64_PATH"}'
```

#### Download File

```bash
curl -X POST http://SERVER:48891/api/file/download \
  -H "Content-Type: application/json" \
  -d '{"path": "BASE64_PATH"}'
```

#### Delete File

```bash
curl -X POST http://SERVER:48891/api/file/delete \
  -H "Content-Type: application/json" \
  -d '{"path": "BASE64_PATH"}'
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

### Example 1: Deploy Project Folder

```bash
# 1. 本地压缩项目
tar -czvf dist.tar.gz ./dist

# 2. 编码目标路径
echo -n "/www/wwwroot/app/" | base64
# L3d3dy93d3dyb290L2FwcC8=

# 3. 上传并自动解压
curl -X POST http://SERVER:48891/api/file/upload \
  -F "file=@dist.tar.gz" \
  -F "path=L3d3dy93d3dyb290L2FwcC8=" \
  -F "extract=true"

# 4. 验证文件
curl -X POST http://SERVER:48891/api/file/list \
  -H "Content-Type: application/json" \
  -d '{"path": "L3d3dy93d3dyb290L2FwcC8="}'
```

### Example 2: Restart PM2

```bash
# 1. Encode command
echo -n "pm2 restart all" | base64
# cG0yIHJlc3RhcnQgYWxs

# 2. Execute
curl -X POST http://SERVER:48891/api/ssh/exec \
  -H "Content-Type: application/json" \
  -d '{"command": "cG0yIHJlc3RhcnQgYWxs"}'
```

### Example 3: PowerShell Workflow

```powershell
# 编码路径
$path = [Convert]::ToBase64String([Text.Encoding]::UTF8.GetBytes("/www/wwwroot/app/"))

# 上传并解压
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
  host: "your-server.com"
  port: 22
  user: "username"
  password: "password"
```

---

## Version History

### v1.2.0 (2026-02-09)
- ✨ **New HTTP File Upload API** - Multipart upload, no FTP required
- ✨ **Auto-extract support** - tar.gz, zip, tar
- ✅ Fixed path handling for directories ending with slash
- ✅ Added comprehensive debug logging

### v1.1.0 (2026-02-08)
- ✅ Fixed SSH connection race condition
- ✅ Fixed WebSocket goroutine leak
- ✅ Added graceful shutdown (SIGINT/SIGTERM)

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
