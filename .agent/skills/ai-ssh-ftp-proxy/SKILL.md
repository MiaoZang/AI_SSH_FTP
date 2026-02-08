---
name: ai-ssh-ftp-proxy
description: "AI Agent Skill for executing SSH commands and FTP operations on remote servers via a proxy service. Supports HTTP API and WebSocket for interactive sessions."
version: "1.1.0"
---

# AI SSH/FTP Proxy Skill

Enable AI agents to securely access remote servers via SSH and FTP through a dedicated proxy service.

## Overview

This skill provides a proxy service that AI agents can call to:
- Execute shell commands on remote servers (SSH)
- Transfer files to/from remote servers (FTP)
- Open interactive shell sessions (WebSocket)

All inputs and outputs are Base64 encoded for safe transmission.

## Quick Start

### One-Line Installation

```bash
curl -fsSL https://raw.githubusercontent.com/MiaoZang/AI_SSH_FTP/main/scripts/manage.sh | bash
```

Or download and run the management script:
```bash
wget https://github.com/MiaoZang/AI_SSH_FTP/releases/latest/download/manage.sh
chmod +x manage.sh
./manage.sh
```

### Management Script Features

- 🌐 **Bilingual** - English and Chinese support
- 📦 **Auto-download** - Downloads binary from GitHub if missing
- 🔧 **Interactive config** - Creates config.yaml via wizard
- 🖥️ **Multi-distro** - Ubuntu/Debian/CentOS/Fedora/Arch/Alpine

## API Endpoints

### SSH Command Execution

```bash
curl -X POST http://YOUR_SERVER:48891/api/ssh/exec \
  -H "Content-Type: application/json" \
  -d '{"command": "BASE64_ENCODED_COMMAND"}'
```

Response:
```json
{
  "stdout": "BASE64_ENCODED_OUTPUT",
  "stderr": "BASE64_ENCODED_ERRORS",
  "exit_code": 0
}
```

### FTP Operations

**List Directory:**
```bash
curl -X POST http://YOUR_SERVER:48891/api/ftp/list \
  -d '{"path": "BASE64_ENCODED_PATH"}'
```

**Upload File:**
```bash
curl -X POST http://YOUR_SERVER:48891/api/ftp/upload \
  -d '{"path": "BASE64_ENCODED_PATH", "content": "BASE64_ENCODED_CONTENT"}'
```

**Download File:**
```bash
curl -X POST http://YOUR_SERVER:48891/api/ftp/download \
  -d '{"path": "BASE64_ENCODED_PATH"}'
```

### WebSocket Interactive SSH

Connect to `ws://YOUR_SERVER:48892/ws/ssh`

**Client → Server:**
```json
{"type": "input", "payload": "BASE64_ENCODED_INPUT"}
```

**Server → Client:**
```json
{"type": "output", "payload": "BASE64_ENCODED_OUTPUT"}
```

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

ftp_server:
  host: "your-server.com"
  port: 21
  user: "username"
  password: "password"
```

## Base64 Examples

```bash
# Encode
echo -n "ls -la /" | base64
# Output: bHMgLWxhIC8=

# Decode
echo "cm9vdAo=" | base64 -d
# Output: root
```

## Version History

### v1.1.0 (2026-02-08)
- ✅ Fixed SSH connection race condition
- ✅ Fixed WebSocket goroutine leak
- ✅ Added graceful shutdown (SIGINT/SIGTERM)
- ✅ Fixed WS error message Base64 encoding
- ✅ Updated management script with bilingual support

### v1.0.0 (2026-01-18)
- Initial release

## Repository

**GitHub**: https://github.com/MiaoZang/AI_SSH_FTP

## Security Notes

- Credentials are stored server-side, not passed per-request
- All data is Base64 encoded to prevent injection attacks
- Deploy on a private network or use firewall rules
- Consider using SSH key authentication over passwords
