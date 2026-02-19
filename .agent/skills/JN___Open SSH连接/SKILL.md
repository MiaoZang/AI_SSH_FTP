---
name: SSH 直连方式
description: 使用 Windows OpenSSH 连接远程服务器并执行命令的通用技能
---

# SSH 直连方式

> 使用 Windows 自带的 OpenSSH 连接远程 Linux 服务器

## 适用场景

- 连接远程 Linux/Unix 服务器
- 执行远程命令、部署应用
- 文件上传下载（SCP）
- 服务器运维操作

---

## ⚠️ 核心规则

> [!CAUTION]
> **保持连接复用！**
> 1. 连接成功后，**保存 CommandId**
> 2. 后续操作使用 `send_command_input` 复用连接
> 3. **禁止每次操作都重新登录**
> 4. 只有连接断开时才可重新登录
> 5. 如果还没有得知ssh服务器的ip，用户名，密码,则先要询问用户获取

---

## 连接流程

### 步骤 1：启动 SSH 连接

```powershell
C:\Windows\System32\OpenSSH\ssh.exe {用户名}@{服务器IP}
```

| 参数 | 说明 | 示例 |
|------|------|------|
| `{用户名}` | SSH登录账号 | root |
| `{服务器IP}` | 服务器地址 | 192.168.1.100 |

### 步骤 2：输入密码

当输出包含 `password:` 时，发送密码：

```typescript
send_command_input({
    CommandId: "{保存的命令ID}",
    Input: "{密码}\n",
    SafeToAutoRun: false,  // 密码必须用户确认
    WaitMs: 5000
})
```

### 步骤 3：执行命令

登录成功后，继续发送命令：

```typescript
send_command_input({
    CommandId: "{保存的命令ID}",
    Input: "{要执行的命令}\n",
    SafeToAutoRun: true,
    WaitMs: 3000
})
```

---

## SCP 文件传输

### 上传文件到服务器

```powershell
C:\Windows\System32\OpenSSH\scp.exe "{本地文件路径}" {用户名}@{服务器IP}:{远程目录}
```

### 从服务器下载文件

```powershell
C:\Windows\System32\OpenSSH\scp.exe {用户名}@{服务器IP}:{远程文件路径} "{本地目录}"
```

---

## AI 操作清单

### 首次连接

```
[ ] 1. 获取连接信息（IP、用户名、密码）
[ ] 2. run_command 启动 SSH 连接
[ ] 3. 等待 password 提示出现
[ ] 4. send_command_input 发送密码（SafeToAutoRun: false）
[ ] 5. 确认登录成功（看到命令提示符如 root@xxx:~#）
[ ] 6. 保存 CommandId 用于后续操作
```

### 执行命令

```
[ ] 1. 使用已保存的 CommandId
[ ] 2. send_command_input 发送命令
[ ] 3. 等待并读取输出
[ ] 4. 检查是否有错误
```

### 连接断开后

```
[ ] 1. 检测到连接已断开
[ ] 2. 通知用户需要重新连接
[ ] 3. 重新执行首次连接流程
```

---

## 常用命令参考

### 文件与目录

| 命令 | 说明 |
|------|------|
| `ls -la` | 列出详细文件列表 |
| `cd {目录}` | 切换目录 |
| `pwd` | 显示当前路径 |
| `mkdir -p {目录}` | 创建目录 |
| `rm -rf {路径}` | 删除文件/目录 |
| `cat {文件}` | 查看文件内容 |

### 进程管理

| 命令 | 说明 |
|------|------|
| `ps aux \| grep {关键词}` | 查找进程 |
| `pkill -f {关键词}` | 杀死进程 |
| `nohup {命令} > log.txt 2>&1 &` | 后台运行 |

### 服务管理

| 命令 | 说明 |
|------|------|
| `systemctl status {服务名}` | 查看服务状态 |
| `systemctl start {服务名}` | 启动服务 |
| `systemctl stop {服务名}` | 停止服务 |
| `systemctl restart {服务名}` | 重启服务 |

### 网络与端口

| 命令 | 说明 |
|------|------|
| `netstat -tlnp` | 查看监听端口 |
| `curl http://localhost:{端口}` | 测试本地服务 |

---

## 注意事项

1. **密码安全**：发送密码时必须 `SafeToAutoRun: false`
2. **连接超时**：长时间无操作可能自动断开
3. **路径格式**：Linux 使用 `/` 分隔，区分大小写
4. **权限问题**：某些操作需要 root 权限或 sudo
