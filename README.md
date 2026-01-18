# AI SSH/FTP Proxy Service

一个为 AI Agent 设计的 SSH/FTP 代理服务，提供 HTTP API 和 WebSocket 接口。

## 功能特性

- **SSH 命令执行**: 通过 HTTP API 执行远程 SSH 命令
- **FTP 文件操作**: 列表、上传、下载文件
- **WebSocket 交互**: 实时交互式 SSH Shell
- **Base64 编码**: 所有输入输出安全编码
- **多语言支持**: 中英文双语管理界面
- **一键部署**: 自动下载、配置向导

## 快速开始

### 方式一：使用管理脚本 (推荐)

```bash
# 下载管理脚本
curl -fsSL https://raw.githubusercontent.com/MiaoZang/AI_SSH_FTP/main/scripts/manage.sh -o manage.sh
chmod +x manage.sh

# 运行 (会自动下载二进制文件并引导配置)
./manage.sh
```

### 方式二：手动编译

```bash
# 克隆仓库
git clone https://github.com/MiaoZang/AI_SSH_FTP.git
cd AI_SSH_FTP

# 编译
go build -o ssh-ftp-proxy cmd/server/main.go

# 配置
cp config/config.yaml.example config/config.yaml
# 编辑 config/config.yaml 填入你的服务器信息

# 启动
./scripts/manage.sh start
```

## API 接口

### SSH 命令执行

```bash
curl -X POST http://localhost:48891/api/ssh/exec \
  -H "Content-Type: application/json" \
  -d '{"command": "BASE64_ENCODED_COMMAND"}'
```

### FTP 操作

```bash
# 列表
curl -X POST http://localhost:48891/api/ftp/list \
  -d '{"path": "BASE64_ENCODED_PATH"}'

# 上传
curl -X POST http://localhost:48891/api/ftp/upload \
  -d '{"path": "BASE64_PATH", "content": "BASE64_CONTENT"}'

# 下载
curl -X POST http://localhost:48891/api/ftp/download \
  -d '{"path": "BASE64_PATH"}'
```

### WebSocket 交互

连接 `ws://localhost:48892/ws/ssh` 进行实时 Shell 交互。

## 配置说明

详见 `config/config.yaml.example`

## License

MIT
