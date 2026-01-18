# SSH 技能手册

使用 SSH 进行安全远程访问、文件传输和隧道转发。

## 基本连接

连接到服务器：
```bash
ssh user@hostname
```

指定端口连接：
```bash
ssh -p 2222 user@hostname
```

使用指定密钥连接：
```bash
ssh -i ~/.ssh/my_key user@hostname
```

## SSH 配置文件

配置文件位置：
```
~/.ssh/config
```

配置文件示例：
```
Host myserver
    HostName 192.168.1.100
    User deploy
    Port 22
    IdentityFile ~/.ssh/myserver_key
    ForwardAgent yes
```

配置后只需执行：
```bash
ssh myserver
```

## 远程执行命令

执行单条命令：
```bash
ssh user@host "ls -la /var/log"
```

执行多条命令：
```bash
ssh user@host "cd /app && git pull && pm2 restart all"
```

交互式运行（需要伪终端）：
```bash
ssh -t user@host "htop"
```

## 使用 SCP 传输文件

上传文件到远程：
```bash
scp local.txt user@host:/remote/path/
```

从远程下载文件：
```bash
scp user@host:/remote/file.txt ./local/
```

递归复制目录：
```bash
scp -r ./local_dir user@host:/remote/path/
```

## 使用 rsync 传输文件（推荐）

同步目录到远程：
```bash
rsync -avz ./local/ user@host:/remote/path/
```

从远程同步：
```bash
rsync -avz user@host:/remote/path/ ./local/
```

显示进度并压缩传输：
```bash
rsync -avzP ./local/ user@host:/remote/path/
```

先执行模拟运行（dry run）：
```bash
rsync -avzn ./local/ user@host:/remote/path/
```

## 端口转发（隧道）

本地转发（在本地访问远程服务）：
```bash
ssh -L 8080:localhost:80 user@host
# 现在 localhost:8080 连接到远程主机的 80 端口
```

本地转发到另一台主机：
```bash
ssh -L 5432:db-server:5432 user@jumphost
# 通过 localhost:5432 访问 db-server:5432
```

远程转发（将本地服务暴露给远程）：
```bash
ssh -R 9000:localhost:3000 user@host
# 远程主机的 9000 端口连接到你本地的 3000 端口
```

动态 SOCKS 代理：
```bash
ssh -D 1080 user@host
# 使用 localhost:1080 作为 SOCKS5 代理
```

## 跳板机 / 堡垒机

通过跳板机连接：
```bash
ssh -J jumphost user@internal-server
```

多级跳转：
```bash
ssh -J jump1,jump2 user@internal-server
```

在配置文件中设置：
```
Host internal
    HostName 10.0.0.50
    User deploy
    ProxyJump bastion
```

## 密钥管理

生成新密钥（Ed25519，推荐）：
```bash
ssh-keygen -t ed25519 -C "your_email@example.com"
```

生成 RSA 密钥（兼容旧系统）：
```bash
ssh-keygen -t rsa -b 4096 -C "your_email@example.com"
```

复制公钥到服务器：
```bash
ssh-copy-id user@host
```

复制指定公钥：
```bash
ssh-copy-id -i ~/.ssh/mykey.pub user@host
```

## SSH 代理（Agent）

启动代理：
```bash
eval "$(ssh-agent -s)"
```

添加密钥到代理：
```bash
ssh-add ~/.ssh/id_ed25519
```

macOS 使用钥匙串：
```bash
ssh-add --apple-use-keychain ~/.ssh/id_ed25519
```

列出已加载的密钥：
```bash
ssh-add -l
```

## 连接复用（Multiplexing）

在 ~/.ssh/config 中配置：
```
Host *
    ControlMaster auto
    ControlPath ~/.ssh/sockets/%r@%h-%p
    ControlPersist 600
```

创建 socket 目录：
```bash
mkdir -p ~/.ssh/sockets
```

## 已知主机管理

移除旧的主机密钥：
```bash
ssh-keygen -R hostname
```

扫描并添加主机密钥：
```bash
ssh-keyscan hostname >> ~/.ssh/known_hosts
```

## 调试连接

详细输出：
```bash
ssh -v user@host
```

更详细输出：
```bash
ssh -vv user@host
```

最大详细程度：
```bash
ssh -vvv user@host
```

## 安全建议

- 使用 Ed25519 密钥（比 RSA 更快更安全）
- 在服务器上设置 `PasswordAuthentication no` 禁用密码登录
- 使用 `fail2ban` 阻止暴力破解
- 使用密码短语加密密钥
- 使用 `ssh-agent` 避免重复输入密码短语
- 在 authorized_keys 中使用 `command=` 限制密钥用途
