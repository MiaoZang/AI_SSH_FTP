#!/bin/bash
# ============================================================
# AI SSH/FTP Proxy - Non-Interactive Install Script
# For AI automated deployment (zero interaction required)
# Also supports human CLI usage with parameters
# ============================================================

set -e

# --- Default Values ---
INSTALL_DIR="/opt/ssh-ftp-proxy"
HTTP_PORT=48891
WS_PORT=48892
BIND_IP="0.0.0.0"
SSH_HOST="127.0.0.1"
SSH_PORT=22
SSH_USER="root"
SSH_PASS=""
SSH_KEYFILE=""
FTP_HOST=""
FTP_PORT=21
FTP_USER=""
FTP_PASS=""
LOG_LEVEL="info"
AUTO_START=false
SETUP_SYSTEMD=false
FORCE_REINSTALL=false

# GitHub
GITHUB_REPO="MiaoZang/AI_SSH_FTP"
GITHUB_RELEASE_URL="https://github.com/$GITHUB_REPO/releases/latest/download"
BINARY_NAME="ssh-ftp-proxy"

# --- Colors ---
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info()    { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[OK]${NC} $1"; }
log_warn()    { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error()   { echo -e "${RED}[ERROR]${NC} $1"; }

# --- Usage ---
usage() {
    cat <<EOF
AI SSH/FTP Proxy - Install Script

Usage: $0 [OPTIONS]

Required:
  --ssh-pass PASSWORD       SSH password for the target server

Server Options:
  --install-dir DIR         Install directory (default: /opt/ssh-ftp-proxy)
  --http-port PORT          HTTP API port (default: 48891)
  --ws-port PORT            WebSocket port (default: 48892)
  --bind-ip IP              Bind IP address (default: 0.0.0.0)

SSH Configuration:
  --ssh-host HOST           SSH host (default: 127.0.0.1)
  --ssh-port PORT           SSH port (default: 22)
  --ssh-user USER           SSH username (default: root)
  --ssh-keyfile FILE        SSH key file (optional)

FTP Configuration (defaults to SSH values):
  --ftp-host HOST           FTP host (default: same as ssh-host)
  --ftp-port PORT           FTP port (default: 21)
  --ftp-user USER           FTP username (default: same as ssh-user)
  --ftp-pass PASSWORD       FTP password (default: same as ssh-pass)

Other:
  --log-level LEVEL         Log level: debug|info|warn|error (default: info)
  --auto-start              Start service after install
  --systemd                 Register as systemd service
  --force                   Force reinstall even if already installed
  -h, --help                Show this help

Examples:
  # Minimal install (AI default)
  $0 --ssh-pass MyPassword --auto-start

  # Full control
  $0 --ssh-host 127.0.0.1 --ssh-port 1643 --ssh-user root --ssh-pass MyPass \\
     --http-port 48891 --auto-start --systemd

  # One-liner from GitHub
  curl -fsSL https://raw.githubusercontent.com/MiaoZang/AI_SSH_FTP/main/scripts/install.sh | \\
    bash -s -- --ssh-pass MyPassword --auto-start
EOF
    exit 0
}

# --- Parse Arguments ---
parse_args() {
    while [[ $# -gt 0 ]]; do
        case "$1" in
            --install-dir)  INSTALL_DIR="$2"; shift 2 ;;
            --http-port)    HTTP_PORT="$2"; shift 2 ;;
            --ws-port)      WS_PORT="$2"; shift 2 ;;
            --bind-ip)      BIND_IP="$2"; shift 2 ;;
            --ssh-host)     SSH_HOST="$2"; shift 2 ;;
            --ssh-port)     SSH_PORT="$2"; shift 2 ;;
            --ssh-user)     SSH_USER="$2"; shift 2 ;;
            --ssh-pass)     SSH_PASS="$2"; shift 2 ;;
            --ssh-keyfile)  SSH_KEYFILE="$2"; shift 2 ;;
            --ftp-host)     FTP_HOST="$2"; shift 2 ;;
            --ftp-port)     FTP_PORT="$2"; shift 2 ;;
            --ftp-user)     FTP_USER="$2"; shift 2 ;;
            --ftp-pass)     FTP_PASS="$2"; shift 2 ;;
            --log-level)    LOG_LEVEL="$2"; shift 2 ;;
            --auto-start)   AUTO_START=true; shift ;;
            --systemd)      SETUP_SYSTEMD=true; shift ;;
            --force)        FORCE_REINSTALL=true; shift ;;
            -h|--help)      usage ;;
            *)
                log_error "Unknown option: $1"
                usage
                ;;
        esac
    done

    # Validate required
    if [[ -z "$SSH_PASS" ]]; then
        log_error "--ssh-pass is required"
        echo ""
        usage
    fi

    # FTP defaults to SSH values
    FTP_HOST=${FTP_HOST:-$SSH_HOST}
    FTP_USER=${FTP_USER:-$SSH_USER}
    FTP_PASS=${FTP_PASS:-$SSH_PASS}
}

# --- Check if already installed ---
check_existing() {
    if [[ -f "$INSTALL_DIR/$BINARY_NAME" ]] && [[ "$FORCE_REINSTALL" != "true" ]]; then
        log_warn "Already installed at $INSTALL_DIR"
        log_info "Use --force to reinstall"

        # If auto-start requested, just start the existing install
        if [[ "$AUTO_START" == "true" ]]; then
            start_service
        fi
        show_summary
        exit 0
    fi
}

# --- Install curl if missing ---
ensure_curl() {
    if command -v curl &> /dev/null; then
        return 0
    fi

    log_info "Installing curl..."
    if command -v apt-get &> /dev/null; then
        apt-get update -qq && apt-get install -y -qq curl
    elif command -v dnf &> /dev/null; then
        dnf install -y -q curl
    elif command -v yum &> /dev/null; then
        yum install -y -q curl
    elif command -v apk &> /dev/null; then
        apk add --quiet curl
    else
        log_error "Cannot install curl. Please install it manually."
        exit 1
    fi
    log_success "curl installed"
}

# --- Create install directory ---
setup_directory() {
    log_info "Setting up directory: $INSTALL_DIR"
    mkdir -p "$INSTALL_DIR"
    mkdir -p "$INSTALL_DIR/config"
    mkdir -p "$INSTALL_DIR/logs"
    log_success "Directory ready"
}

# --- Download binary ---
download_binary() {
    log_info "Downloading binary from GitHub..."
    local url="$GITHUB_RELEASE_URL/$BINARY_NAME"

    if curl -fsSL -o "$INSTALL_DIR/$BINARY_NAME" "$url" 2>/dev/null; then
        chmod +x "$INSTALL_DIR/$BINARY_NAME"
        local size=$(du -h "$INSTALL_DIR/$BINARY_NAME" | cut -f1)
        log_success "Binary downloaded ($size)"
        return 0
    fi

    log_error "Download failed from: $url"
    log_info "Please check network connection or download manually"
    exit 1
}

# --- Generate config ---
generate_config() {
    local config_file="$INSTALL_DIR/config/config.yaml"

    log_info "Generating config: $config_file"

    cat > "$config_file" <<YAML
server:
  http_port: $HTTP_PORT
  ws_port: $WS_PORT
  bind_ip: "$BIND_IP"

ssh_server:
  host: "$SSH_HOST"
  port: $SSH_PORT
  user: "$SSH_USER"
  password: "$SSH_PASS"
  key_file: "$SSH_KEYFILE"

ftp_server:
  host: "$FTP_HOST"
  port: $FTP_PORT
  user: "$FTP_USER"
  password: "$FTP_PASS"

log:
  level: "$LOG_LEVEL"
  file: "logs/server.log"
YAML

    log_success "Config generated"
}

# --- Setup systemd service ---
setup_systemd_service() {
    if [[ "$SETUP_SYSTEMD" != "true" ]]; then
        return 0
    fi

    log_info "Creating systemd service..."

    cat > /etc/systemd/system/ssh-ftp-proxy.service <<SERVICE
[Unit]
Description=AI SSH/FTP Proxy Service
After=network.target

[Service]
Type=simple
WorkingDirectory=$INSTALL_DIR
ExecStart=$INSTALL_DIR/$BINARY_NAME
Restart=on-failure
RestartSec=5
StandardOutput=append:$INSTALL_DIR/logs/server.log
StandardError=append:$INSTALL_DIR/logs/server.log

[Install]
WantedBy=multi-user.target
SERVICE

    systemctl daemon-reload
    systemctl enable ssh-ftp-proxy
    log_success "Systemd service created and enabled"
}

# --- Start service ---
start_service() {
    log_info "Starting service..."

    # Stop existing if running
    pkill -f "$INSTALL_DIR/$BINARY_NAME" 2>/dev/null || true
    sleep 1

    if [[ "$SETUP_SYSTEMD" == "true" ]]; then
        systemctl start ssh-ftp-proxy
        sleep 2
        if systemctl is-active --quiet ssh-ftp-proxy; then
            local pid=$(systemctl show ssh-ftp-proxy --property=MainPID --value)
            log_success "Service started via systemd (PID: $pid)"
        else
            log_error "Service failed to start"
            systemctl status ssh-ftp-proxy --no-pager || true
            exit 1
        fi
    else
        cd "$INSTALL_DIR"
        nohup ./$BINARY_NAME > logs/server.log 2>&1 &
        local pid=$!
        echo "$pid" > "$INSTALL_DIR/server.pid"
        sleep 2

        if kill -0 "$pid" 2>/dev/null; then
            log_success "Service started (PID: $pid)"
        else
            log_error "Service failed to start. Check logs:"
            tail -10 "$INSTALL_DIR/logs/server.log" 2>/dev/null || true
            exit 1
        fi
    fi
}

# --- Health check ---
health_check() {
    log_info "Running health check..."
    sleep 1

    local url="http://127.0.0.1:$HTTP_PORT/api/health"
    local max_retries=5
    local retry=0

    while [[ $retry -lt $max_retries ]]; do
        if curl -fsSL "$url" 2>/dev/null | grep -q "ok"; then
            log_success "Health check passed: $url"
            return 0
        fi
        retry=$((retry + 1))
        sleep 1
    done

    log_warn "Health check failed after $max_retries retries"
    log_info "Service may still be starting. Check: curl $url"
}

# --- Print summary ---
show_summary() {
    echo ""
    echo -e "${GREEN}============================================================${NC}"
    echo -e "${GREEN}  INSTALLATION COMPLETE${NC}"
    echo -e "${GREEN}============================================================${NC}"
    echo ""
    echo -e "  Install Dir:    ${BLUE}$INSTALL_DIR${NC}"
    echo -e "  HTTP API:       ${BLUE}http://0.0.0.0:$HTTP_PORT${NC}"
    echo -e "  WebSocket:      ${BLUE}ws://0.0.0.0:$WS_PORT${NC}"
    echo -e "  Health Check:   ${BLUE}http://127.0.0.1:$HTTP_PORT/api/health${NC}"
    echo -e "  Config:         ${BLUE}$INSTALL_DIR/config/config.yaml${NC}"
    echo -e "  Logs:           ${BLUE}$INSTALL_DIR/logs/server.log${NC}"
    echo ""
    echo -e "  SSH Target:     ${BLUE}$SSH_USER@$SSH_HOST:$SSH_PORT${NC}"
    if [[ "$SETUP_SYSTEMD" == "true" ]]; then
        echo -e "  Systemd:        ${BLUE}systemctl {start|stop|restart|status} ssh-ftp-proxy${NC}"
    fi
    echo ""
    echo -e "${GREEN}============================================================${NC}"
}

# --- Main ---
main() {
    echo ""
    echo -e "${BLUE}============================================================${NC}"
    echo -e "${BLUE}  AI SSH/FTP Proxy - Installer${NC}"
    echo -e "${BLUE}============================================================${NC}"
    echo ""

    parse_args "$@"
    check_existing
    ensure_curl
    setup_directory
    download_binary
    generate_config
    setup_systemd_service

    if [[ "$AUTO_START" == "true" ]]; then
        start_service
        health_check
    fi

    show_summary
}

main "$@"
