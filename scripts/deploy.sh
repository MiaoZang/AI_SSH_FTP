#!/bin/bash
# ============================================================
# AI SSH/FTP Proxy - 一键项目部署脚本
# One-Click Project Deployment Script
# ============================================================
# 用法 / Usage:
#   ./deploy.sh <LOCAL_DIR> <REMOTE_PATH> [SERVER_URL]
#
# 示例 / Examples:
#   ./deploy.sh ./dist /www/wwwroot/app/
#   ./deploy.sh ./src /home/user/project/ http://myserver:48891
# ============================================================

set -e

# --- Colors ---
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

print_success() { echo -e "${GREEN}✓ $1${NC}"; }
print_error() { echo -e "${RED}✗ $1${NC}"; }
print_warning() { echo -e "${YELLOW}⚠ $1${NC}"; }
print_info() { echo -e "${BLUE}ℹ $1${NC}"; }

# --- Configuration ---
DEFAULT_SERVER="http://127.0.0.1:48891"
TEMP_ARCHIVE="/tmp/deploy_$(date +%s).tar.gz"

# --- Parse Arguments ---
LOCAL_DIR="${1:-.}"
REMOTE_PATH="${2:-/tmp/deploy/}"
SERVER_URL="${3:-$DEFAULT_SERVER}"

# --- Validation ---
if [[ ! -d "$LOCAL_DIR" ]]; then
    print_error "本地目录不存在: $LOCAL_DIR"
    exit 1
fi

# Resolve absolute path
LOCAL_DIR=$(cd "$LOCAL_DIR" && pwd)
DIR_NAME=$(basename "$LOCAL_DIR")

# --- Banner ---
echo ""
echo -e "${BLUE}╔════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║          AI SSH/FTP Proxy - 一键部署 / One-Click Deploy    ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════════════════════╝${NC}"
echo ""
echo "  本地目录 / Local:    $LOCAL_DIR"
echo "  远程路径 / Remote:   $REMOTE_PATH"
echo "  服务地址 / Server:   $SERVER_URL"
echo ""

# --- Step 1: Health Check ---
print_info "[1/4] 检查服务状态 / Checking server..."
HEALTH_RESPONSE=$(curl -s --max-time 5 "$SERVER_URL/api/health" 2>/dev/null || echo "")
if [[ "$HEALTH_RESPONSE" != *"ok"* ]]; then
    print_error "无法连接到服务器 / Cannot connect to server: $SERVER_URL"
    exit 1
fi
print_success "服务正常 / Server OK"

# --- Step 2: Compress ---
print_info "[2/4] 压缩目录 / Compressing..."
cd "$(dirname "$LOCAL_DIR")"
tar -czf "$TEMP_ARCHIVE" "$DIR_NAME" 2>/dev/null
ARCHIVE_SIZE=$(ls -lh "$TEMP_ARCHIVE" | awk '{print $5}')
print_success "压缩完成 / Compressed: $ARCHIVE_SIZE"

# --- Step 3: Upload & Extract ---
print_info "[3/4] 上传并解压 / Uploading & extracting..."

# Base64 encode remote path
if command -v base64 &> /dev/null; then
    PATH_B64=$(echo -n "$REMOTE_PATH" | base64 -w 0 2>/dev/null || echo -n "$REMOTE_PATH" | base64)
else
    print_error "需要 base64 命令 / base64 command required"
    exit 1
fi

# Upload with curl
UPLOAD_RESPONSE=$(curl -s --max-time 300 \
    -X POST "$SERVER_URL/api/file/upload" \
    -F "file=@$TEMP_ARCHIVE" \
    -F "path=$PATH_B64" \
    -F "extract=true" 2>/dev/null || echo '{"error": "upload failed"}')

# Check response
if [[ "$UPLOAD_RESPONSE" == *'"success":true'* ]]; then
    print_success "上传成功 / Upload successful"
else
    print_error "上传失败 / Upload failed"
    echo "Response: $UPLOAD_RESPONSE"
    rm -f "$TEMP_ARCHIVE"
    exit 1
fi

# --- Step 4: Cleanup ---
rm -f "$TEMP_ARCHIVE"

# --- Step 5: Verify ---
print_info "[4/4] 验证部署 / Verifying..."
LIST_RESPONSE=$(curl -s --max-time 10 \
    -X POST "$SERVER_URL/api/file/list" \
    -H "Content-Type: application/json" \
    -d "{\"path\": \"$PATH_B64\"}" 2>/dev/null || echo '{"error": "list failed"}')

if [[ "$LIST_RESPONSE" == *'"files":'* ]]; then
    print_success "部署完成 / Deployment complete"
    echo ""
    echo -e "${GREEN}═══════════════════════════════════════════════════════════════${NC}"
    echo -e "${GREEN}   ✓ 部署成功 / Deployment Successful${NC}"
    echo -e "${GREEN}     远程路径 / Remote: $REMOTE_PATH$DIR_NAME/${NC}"
    echo -e "${GREEN}═══════════════════════════════════════════════════════════════${NC}"
else
    print_warning "无法验证，但上传成功 / Cannot verify, but upload succeeded"
fi

echo ""
