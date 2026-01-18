---
name: ai-ssh-ftp-proxy
description: "AI Agent Skill for executing SSH commands and FTP operations on remote servers via a proxy service. Supports HTTP API and WebSocket for interactive sessions."
---

# AI SSH/FTP Proxy Skill

Enable AI agents to securely access remote servers via SSH and FTP through a dedicated proxy service.

## Overview

This skill provides a proxy service that AI agents can call to:
- Execute shell commands on remote servers (SSH)
- Transfer files to/from remote servers (FTP)
- Open interactive shell sessions (WebSocket)

All inputs and outputs are Base64 encoded for safe transmission.

## Deployment

1. Upload the binary and config to your server
2. Configure SSH/FTP credentials in `config/config.yaml`
3. Start with `./scripts/start.sh`

## API Endpoints

### SSH Command Execution

Execute a command on the remote server:
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
  -H "Content-Type: application/json" \
  -d '{"path": "BASE64_ENCODED_PATH"}'
```

**Upload File:**
```bash
curl -X POST http://YOUR_SERVER:48891/api/ftp/upload \
  -H "Content-Type: application/json" \
  -d '{"path": "BASE64_ENCODED_PATH", "content": "BASE64_ENCODED_CONTENT"}'
```

**Download File:**
```bash
curl -X POST http://YOUR_SERVER:48891/api/ftp/download \
  -H "Content-Type: application/json" \
  -d '{"path": "BASE64_ENCODED_PATH"}'
```

### WebSocket Interactive SSH

Connect to `ws://YOUR_SERVER:48892/ws/ssh` for an interactive shell session.

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

## Base64 Encoding Examples

Encode a command:
```bash
echo -n "ls -la /" | base64
# Output: bHMgLWxhIC8=
```

Decode a response:
```bash
echo "dG90YWwgOTYK..." | base64 -d
```

## Security Notes

- Credentials are stored server-side, not passed per-request
- All data is Base64 encoded to prevent injection attacks
- Deploy on a private network or use firewall rules
- Consider using SSH key authentication over passwords
