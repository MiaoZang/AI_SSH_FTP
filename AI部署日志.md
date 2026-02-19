现在我明白了！AI SSH FTP 代理支持一键安装。我需要：

先通过 OpenSSH 直连到服务器
在服务器上运行一键安装脚本
配置好 SSH 连接信息（连接到自身 localhost）
让我开始操作：

3
通过 OpenSSH 直连服务器并安装 AI SSH FTP 代理

…\C++_save\paypro > C:\Windows\System32\OpenSSH\ssh.exe -o StrictHostKeyChecking=no -p 1643 root@154.222.29.186

root@gsdIQ8D1FQxlo02giI:~# bt 5 2>&1 | head -20
===============================================
正在执行(5)...
===============================================
ss -tlnp | grep 48891 && echo "PORT LISTENING" || echo "PORT NO
T FOUND"
请输入新的面板密码：|-用户名: swfo1b8k
|-新密码: ss -tlnp | grep 48891 && echo "PORT LISTENING" || ech
o "PORT NOT FOUND"
root@gsdIQ8D1FQxlo02giI:~#  
Running
Relocate
Cancel
SSH 连接已建立，等待密码输入：

Sent input to command
MvGa84lJb0J4↵
SSH 连接成功！现在安装 AI SSH FTP 代理服务。先下载并运行一键安装脚本：

Sent input to command
curl -fsSL https://raw.githubusercontent.com/MiaoZang/AI_SSH_FTP/main/scripts/manage.sh | bash↵
看起来安装脚本显示了语言选择菜单但直接退出了（因为用管道方式执行无法交互）。让我先下载脚本再手动执行：

Sent input to command
curl -fsSL https://raw.githubusercontent.com/MiaoZang/AI_SSH_FTP/main/scripts/manage.sh -o /tmp/manage.sh && chmod +x /tmp/manage.sh && /tmp/manage.sh↵
需要选择语言。选择中文：

Sent input to command
2↵
安装正在进行，正在下载程序。让我等待下载完成：

Checked command status
下载成功，开始配置向导了。我需要输入配置信息。按默认端口 48891 即可：

4
完成 AI SSH FTP 代理的配置向导

回车使用默认 HTTP 端口，然后继续配置：

Sent input to command
↵
WebSocket 端口也用默认的。继续回车：

Sent input to command
↵
绑定 IP 用默认。回车继续，然后需要输入 SSH 连接信息（指向自身 localhost）：

Sent input to command
↵
需要输入 SSH 主机。由于代理运行在这台服务器本地，SSH 连接的是自身 localhost，端口是 1643：

Sent input to command
127.0.0.1↵
SSH 端口需要改为 1643：

Sent input to command
1643↵
5
输入 SSH 用户名、密码并完成配置向导启动服务

输入 SSH 用户名 root：

Sent input to command
root↵
输入 SSH 密码：

Sent input to command
MvGa84lJb0J4↵
SSH 密钥文件跳过：

Sent input to command
↵
FTP 配置也直接默认即可：

Sent input to command
↵
Sent input to command
↵
FTP 用户名用默认 root：

Sent input to command
↵
FTP 密码自动填入了相同密码，直接回车确认：

6
完成 FTP 配置向导并启动代理服务

Sent input to command
↵
日志级别默认 debug 即可：

Sent input to command
↵
配置完成了！管理菜单出现了。选择 1 以启动服务：

Sent input to command
1↵
服务已启动成功，PID 66908。让我按回车退出管理菜单，然后验证服务是否正常工作：

Sent input to command
↵
让我退出管理菜单，然后通过 HTTP API 测试服务是否正常：

Sent input to command
0↵
好的，退出管理菜单成功了。现在通过 HTTP API 验证 AI SSH FTP 代理服务是否正常工作：

7
验证 AI SSH FTP 代理服务是否正常响应