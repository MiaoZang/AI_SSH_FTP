#!/bin/bash
# ============================================================
# AI SSH/FTP Proxy Service - Unified Management Script
# 支持中英文双语 / Bilingual Support (Chinese & English)
# 支持多发行版 / Multi-Distro Support
# ============================================================

set -e

# --- Configuration ---
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$SCRIPT_DIR/.." || exit 1

BINARY_NAME="ssh-ftp-proxy"
CONFIG_FILE="config/config.yaml"
LOG_FILE="config/server.log"
PID_FILE="server.pid"
LANG_FILE=".lang_preference"

# GitHub Repository
GITHUB_REPO="MiaoZang/AI_SSH_FTP"
GITHUB_RAW_URL="https://raw.githubusercontent.com/$GITHUB_REPO/main"
GITHUB_RELEASE_URL="https://github.com/$GITHUB_REPO/releases/latest/download"

# --- Colors ---
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# --- Distro Detection ---
detect_distro() {
    if [[ -f /etc/os-release ]]; then
        . /etc/os-release
        DISTRO=$ID
        DISTRO_VERSION=$VERSION_ID
    elif [[ -f /etc/redhat-release ]]; then
        DISTRO="rhel"
    elif [[ -f /etc/debian_version ]]; then
        DISTRO="debian"
    else
        DISTRO="unknown"
    fi
    echo "$DISTRO"
}

# --- Package Manager ---
get_pkg_manager() {
    if command -v apt-get &> /dev/null; then
        echo "apt"
    elif command -v dnf &> /dev/null; then
        echo "dnf"
    elif command -v yum &> /dev/null; then
        echo "yum"
    elif command -v pacman &> /dev/null; then
        echo "pacman"
    elif command -v apk &> /dev/null; then
        echo "apk"
    elif command -v zypper &> /dev/null; then
        echo "zypper"
    else
        echo "unknown"
    fi
}

install_package() {
    local pkg="$1"
    local pkg_manager=$(get_pkg_manager)
    
    case $pkg_manager in
        apt)
            sudo apt-get update -qq && sudo apt-get install -y -qq "$pkg"
            ;;
        dnf)
            sudo dnf install -y -q "$pkg"
            ;;
        yum)
            sudo yum install -y -q "$pkg"
            ;;
        pacman)
            sudo pacman -Sy --noconfirm "$pkg"
            ;;
        apk)
            sudo apk add --quiet "$pkg"
            ;;
        zypper)
            sudo zypper install -y "$pkg"
            ;;
        *)
            print_error "$(t unknown_pkg_manager)"
            return 1
            ;;
    esac
}

# --- Language Strings ---
LANG="en"

declare -A EN=(
    ["banner_title"]="AI SSH/FTP Proxy Service - Management Console"
    ["select_option"]="Please select an option:"
    ["start"]="Start Service"
    ["stop"]="Stop Service"
    ["restart"]="Restart Service"
    ["status"]="Check Status"
    ["logs"]="View Logs (live)"
    ["reconfig"]="Reconfigure"
    ["lang_switch"]="Switch Language (切换语言)"
    ["update"]="Update Binary"
    ["exit"]="Exit"
    ["enter_choice"]="Enter choice"
    ["press_enter"]="Press Enter to continue..."
    ["invalid_option"]="Invalid option"
    ["checking_deps"]="Checking dependencies..."
    ["binary_not_found"]="Binary not found!"
    ["downloading_binary"]="Downloading binary from GitHub..."
    ["download_success"]="Binary downloaded successfully"
    ["download_failed"]="Failed to download binary"
    ["compile_hint"]="Please compile the binary first with: go build -o"
    ["fixing_permission"]="Binary is not executable. Fixing..."
    ["deps_ok"]="Dependencies check passed"
    ["config_exists"]="Config file exists"
    ["config_not_found"]="Config file not found. Starting configuration wizard..."
    ["config_wizard"]="=== Configuration Wizard ==="
    ["http_port"]="HTTP Port"
    ["ws_port"]="WebSocket Port"
    ["bind_ip"]="Bind IP"
    ["ssh_config"]="--- SSH Server Configuration ---"
    ["ssh_host"]="SSH Host"
    ["ssh_port"]="SSH Port"
    ["ssh_user"]="SSH User"
    ["ssh_password"]="SSH Password"
    ["ssh_keyfile"]="SSH Key File (optional, press Enter to skip)"
    ["ftp_config"]="--- FTP Server Configuration ---"
    ["ftp_host"]="FTP Host"
    ["ftp_port"]="FTP Port"
    ["ftp_user"]="FTP User"
    ["ftp_password"]="FTP Password"
    ["log_level"]="Log Level [debug/info/warn/error]"
    ["config_created"]="Config file created"
    ["already_running"]="Service is already running"
    ["starting"]="Starting service..."
    ["started"]="Service started"
    ["logs_at"]="Logs"
    ["start_failed"]="Failed to start service. Check logs"
    ["not_running"]="Service is not running"
    ["stopping"]="Stopping service..."
    ["force_killing"]="Force killing..."
    ["stopped"]="Service stopped"
    ["running"]="Service is RUNNING"
    ["stopped_status"]="Service is STOPPED"
    ["recent_logs"]="Recent logs:"
    ["no_logs"]="No logs available"
    ["showing_logs"]="Showing logs (Ctrl+C to exit)..."
    ["no_log_file"]="No log file found"
    ["created_config_dir"]="Created config directory"
    ["installing_curl"]="Installing curl..."
    ["distro_detected"]="Detected distro"
    ["unknown_pkg_manager"]="Unknown package manager, please install dependencies manually"
    ["updating_binary"]="Updating binary from GitHub..."
)

declare -A ZH=(
    ["banner_title"]="AI SSH/FTP 代理服务 - 管理控制台"
    ["select_option"]="请选择操作:"
    ["start"]="启动服务"
    ["stop"]="停止服务"
    ["restart"]="重启服务"
    ["status"]="查看状态"
    ["logs"]="查看日志 (实时)"
    ["reconfig"]="重新配置"
    ["lang_switch"]="切换语言 (Switch Language)"
    ["update"]="更新程序"
    ["exit"]="退出"
    ["enter_choice"]="请输入选项"
    ["press_enter"]="按回车键继续..."
    ["invalid_option"]="无效选项"
    ["checking_deps"]="检查依赖项..."
    ["binary_not_found"]="未找到可执行文件!"
    ["downloading_binary"]="正在从 GitHub 下载程序..."
    ["download_success"]="程序下载成功"
    ["download_failed"]="下载失败"
    ["compile_hint"]="请先编译: go build -o"
    ["fixing_permission"]="可执行文件没有执行权限，正在修复..."
    ["deps_ok"]="依赖检查通过"
    ["config_exists"]="配置文件已存在"
    ["config_not_found"]="未找到配置文件，启动配置向导..."
    ["config_wizard"]="=== 配置向导 ==="
    ["http_port"]="HTTP 端口"
    ["ws_port"]="WebSocket 端口"
    ["bind_ip"]="绑定 IP"
    ["ssh_config"]="--- SSH 服务器配置 ---"
    ["ssh_host"]="SSH 主机"
    ["ssh_port"]="SSH 端口"
    ["ssh_user"]="SSH 用户名"
    ["ssh_password"]="SSH 密码"
    ["ssh_keyfile"]="SSH 密钥文件 (可选，按回车跳过)"
    ["ftp_config"]="--- FTP 服务器配置 ---"
    ["ftp_host"]="FTP 主机"
    ["ftp_port"]="FTP 端口"
    ["ftp_user"]="FTP 用户名"
    ["ftp_password"]="FTP 密码"
    ["log_level"]="日志级别 [debug/info/warn/error]"
    ["config_created"]="配置文件已创建"
    ["already_running"]="服务已在运行中"
    ["starting"]="正在启动服务..."
    ["started"]="服务已启动"
    ["logs_at"]="日志位置"
    ["start_failed"]="启动失败，请检查日志"
    ["not_running"]="服务未运行"
    ["stopping"]="正在停止服务..."
    ["force_killing"]="强制终止..."
    ["stopped"]="服务已停止"
    ["running"]="服务运行中"
    ["stopped_status"]="服务已停止"
    ["recent_logs"]="最近日志:"
    ["no_logs"]="暂无日志"
    ["showing_logs"]="显示日志中 (Ctrl+C 退出)..."
    ["no_log_file"]="日志文件不存在"
    ["created_config_dir"]="已创建配置目录"
    ["installing_curl"]="正在安装 curl..."
    ["distro_detected"]="检测到发行版"
    ["unknown_pkg_manager"]="未知包管理器，请手动安装依赖"
    ["updating_binary"]="正在从 GitHub 更新程序..."
)

t() {
    local key="$1"
    if [[ "$LANG" == "zh" ]]; then
        echo "${ZH[$key]}"
    else
        echo "${EN[$key]}"
    fi
}

# --- Language Selection ---
load_language() {
    if [[ -f "$LANG_FILE" ]]; then
        LANG=$(cat "$LANG_FILE")
    fi
}

save_language() {
    echo "$LANG" > "$LANG_FILE"
}

select_language() {
    echo ""
    echo "╔════════════════════════════════════════╗"
    echo "║     Select Language / 选择语言         ║"
    echo "╠════════════════════════════════════════╣"
    echo "║  1) English                            ║"
    echo "║  2) 中文                               ║"
    echo "╚════════════════════════════════════════╝"
    echo ""
    read -p "Enter choice / 请选择 [1-2]: " lang_choice

    case $lang_choice in
        1) LANG="en" ;;
        2) LANG="zh" ;;
        *) LANG="en" ;;
    esac
    save_language
    print_success "Language set to: $([ "$LANG" == "zh" ] && echo "中文" || echo "English")"
}

# --- Helper Functions ---
print_banner() {
    echo -e "${BLUE}"
    echo "╔════════════════════════════════════════════════════════════╗"
    printf "║ %-58s ║\n" "$(t banner_title)"
    echo "╚════════════════════════════════════════════════════════════╝"
    echo -e "${NC}"
}

print_success() { echo -e "${GREEN}✓ $1${NC}"; }
print_error() { echo -e "${RED}✗ $1${NC}"; }
print_warning() { echo -e "${YELLOW}⚠ $1${NC}"; }
print_info() { echo -e "${BLUE}ℹ $1${NC}"; }

# --- Download Binary ---
download_binary() {
    print_info "$(t downloading_binary)"
    
    # Ensure curl is installed
    if ! command -v curl &> /dev/null; then
        print_warning "$(t installing_curl)"
        install_package curl
    fi

    # Try to download from GitHub releases
    local download_url="$GITHUB_RELEASE_URL/$BINARY_NAME"
    
    if curl -fsSL -o "$BINARY_NAME" "$download_url" 2>/dev/null; then
        chmod +x "$BINARY_NAME"
        print_success "$(t download_success)"
        return 0
    fi

    # Fallback: Clone repo and build (requires Go)
    if command -v go &> /dev/null; then
        print_info "Trying to build from source..."
        local temp_dir=$(mktemp -d)
        if git clone --depth 1 "https://github.com/$GITHUB_REPO.git" "$temp_dir" 2>/dev/null; then
            cd "$temp_dir"
            CGO_ENABLED=0 go build -o "$SCRIPT_DIR/../$BINARY_NAME" cmd/server/main.go
            cd "$SCRIPT_DIR/.."
            rm -rf "$temp_dir"
            chmod +x "$BINARY_NAME"
            print_success "$(t download_success)"
            return 0
        fi
    fi

    print_error "$(t download_failed)"
    print_info "$(t compile_hint) $BINARY_NAME cmd/server/main.go"
    return 1
}

update_binary() {
    print_info "$(t updating_binary)"
    
    # Stop service if running
    if is_running; then
        stop_service
    fi
    
    # Backup old binary
    if [[ -f "$BINARY_NAME" ]]; then
        mv "$BINARY_NAME" "${BINARY_NAME}.bak"
    fi
    
    # Download new binary
    if download_binary; then
        rm -f "${BINARY_NAME}.bak"
        print_success "$(t download_success)"
    else
        # Restore backup
        if [[ -f "${BINARY_NAME}.bak" ]]; then
            mv "${BINARY_NAME}.bak" "$BINARY_NAME"
        fi
        print_error "$(t download_failed)"
    fi
}

# --- Dependency Check ---
check_dependencies() {
    print_info "$(t checking_deps)"
    
    local distro=$(detect_distro)
    print_info "$(t distro_detected): $distro"

    # Check if binary exists, download if not
    if [[ ! -f "$BINARY_NAME" ]]; then
        print_warning "$(t binary_not_found)"
        download_binary || exit 1
    fi

    # Check executable permission
    if [[ ! -x "$BINARY_NAME" ]]; then
        print_warning "$(t fixing_permission)"
        chmod +x "$BINARY_NAME"
    fi

    # Check config directory
    if [[ ! -d "config" ]]; then
        mkdir -p config
        print_success "$(t created_config_dir)"
    fi

    print_success "$(t deps_ok)"
}

# --- Config Wizard ---
check_config() {
    if [[ -f "$CONFIG_FILE" ]]; then
        print_success "$(t config_exists): $CONFIG_FILE"
        return 0
    fi

    print_warning "$(t config_not_found)"
    create_config
}

create_config() {
    echo ""
    print_info "$(t config_wizard)"
    echo ""

    read -p "$(t http_port) [48891]: " http_port
    http_port=${http_port:-48891}

    read -p "$(t ws_port) [48892]: " ws_port
    ws_port=${ws_port:-48892}

    read -p "$(t bind_ip) [0.0.0.0]: " bind_ip
    bind_ip=${bind_ip:-0.0.0.0}

    echo ""
    print_info "$(t ssh_config)"
    read -p "$(t ssh_host): " ssh_host
    read -p "$(t ssh_port) [22]: " ssh_port
    ssh_port=${ssh_port:-22}
    read -p "$(t ssh_user): " ssh_user
    read -s -p "$(t ssh_password): " ssh_password
    echo ""
    read -p "$(t ssh_keyfile): " ssh_keyfile

    echo ""
    print_info "$(t ftp_config)"
    read -p "$(t ftp_host) [$ssh_host]: " ftp_host
    ftp_host=${ftp_host:-$ssh_host}
    read -p "$(t ftp_port) [21]: " ftp_port
    ftp_port=${ftp_port:-21}
    read -p "$(t ftp_user) [$ssh_user]: " ftp_user
    ftp_user=${ftp_user:-$ssh_user}
    read -s -p "$(t ftp_password) [$ssh_password]: " ftp_password
    ftp_password=${ftp_password:-$ssh_password}
    echo ""

    read -p "$(t log_level) [debug]: " log_level
    log_level=${log_level:-debug}

    cat > "$CONFIG_FILE" << EOF
server:
  http_port: $http_port
  ws_port: $ws_port
  bind_ip: "$bind_ip"

ssh_server:
  host: "$ssh_host"
  port: $ssh_port
  user: "$ssh_user"
  password: "$ssh_password"
  key_file: "$ssh_keyfile"

ftp_server:
  host: "$ftp_host"
  port: $ftp_port
  user: "$ftp_user"
  password: "$ftp_password"

log:
  level: "$log_level"
  file: "config/server.log"
EOF

    print_success "$(t config_created): $CONFIG_FILE"
}

# --- Service Control ---
get_pid() {
    if [[ -f "$PID_FILE" ]]; then
        cat "$PID_FILE"
    else
        echo ""
    fi
}

is_running() {
    local pid=$(get_pid)
    if [[ -n "$pid" ]] && kill -0 "$pid" 2>/dev/null; then
        return 0
    else
        return 1
    fi
}

start_service() {
    if is_running; then
        print_warning "$(t already_running) (PID: $(get_pid))"
        return 0
    fi

    print_info "$(t starting)"
    nohup ./"$BINARY_NAME" > "$LOG_FILE" 2>&1 &
    echo $! > "$PID_FILE"
    sleep 1

    if is_running; then
        print_success "$(t started) (PID: $(get_pid))"
        print_info "$(t logs_at): $LOG_FILE"
    else
        print_error "$(t start_failed): $LOG_FILE"
        exit 1
    fi
}

stop_service() {
    if ! is_running; then
        print_warning "$(t not_running)"
        [[ -f "$PID_FILE" ]] && rm -f "$PID_FILE"
        return 0
    fi

    local pid=$(get_pid)
    print_info "$(t stopping) (PID: $pid)..."
    kill "$pid" 2>/dev/null

    local count=0
    while is_running && [[ $count -lt 10 ]]; do
        sleep 1
        ((count++))
    done

    if is_running; then
        print_warning "$(t force_killing)"
        kill -9 "$pid" 2>/dev/null
    fi

    rm -f "$PID_FILE"
    print_success "$(t stopped)"
}

restart_service() {
    stop_service
    sleep 1
    start_service
}

status_service() {
    if is_running; then
        print_success "$(t running) (PID: $(get_pid))"
        echo ""
        print_info "$(t recent_logs)"
        tail -n 10 "$LOG_FILE" 2>/dev/null || echo "$(t no_logs)"
    else
        print_warning "$(t stopped_status)"
    fi
}

view_logs() {
    if [[ -f "$LOG_FILE" ]]; then
        print_info "$(t showing_logs)"
        tail -f "$LOG_FILE"
    else
        print_warning "$(t no_log_file)"
    fi
}

# --- Interactive Menu ---
show_menu() {
    clear
    print_banner
    echo "$(t select_option)"
    echo ""
    echo "  1) $(t start)"
    echo "  2) $(t stop)"
    echo "  3) $(t restart)"
    echo "  4) $(t status)"
    echo "  5) $(t logs)"
    echo "  6) $(t reconfig)"
    echo "  7) $(t update)"
    echo "  8) $(t lang_switch)"
    echo "  0) $(t exit)"
    echo ""
    read -p "$(t enter_choice) [0-8]: " choice

    case $choice in
        1) start_service ;;
        2) stop_service ;;
        3) restart_service ;;
        4) status_service ;;
        5) view_logs ;;
        6) create_config ;;
        7) update_binary ;;
        8) select_language ;;
        0) exit 0 ;;
        *) print_error "$(t invalid_option)" ;;
    esac

    echo ""
    read -p "$(t press_enter)"
    show_menu
}

# --- Main Entry ---
main() {
    load_language

    # First run: ask for language
    if [[ ! -f "$LANG_FILE" ]]; then
        select_language
    fi

    check_dependencies
    check_config

    if [[ $# -eq 0 ]]; then
        show_menu
    else
        case "$1" in
            start) start_service ;;
            stop) stop_service ;;
            restart) restart_service ;;
            status) status_service ;;
            logs) view_logs ;;
            config) create_config ;;
            update) update_binary ;;
            lang) select_language ;;
            *)
                echo "Usage: $0 {start|stop|restart|status|logs|config|update|lang}"
                echo "       $0          (interactive menu)"
                exit 1
                ;;
        esac
    fi
}

main "$@"
