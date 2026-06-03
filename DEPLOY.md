# GoPay 部署指南

本文档指导您在各平台上部署 GoPay 聚合支付网关系统。

## 环境要求

GoPay 已编译为独立静态二进制文件，**无需安装额外依赖**：

- 无需安装 Go、Node.js 或其他运行时
- 数据库使用内嵌 SQLite，无需额外数据库服务
- 最低配置：512MB 内存，1 核 CPU，100MB 磁盘空间

## 快速开始

### 1. 下载

前往 [GitHub Releases](../../releases) 页面，下载对应平台的最新版本：

| 平台 | 文件名 |
|------|--------|
| Linux CLI | `gopay-X.X.X-linux-amd64` |
| Linux GUI（系统托盘）| `gopay-X.X.X-linux-gui-amd64` |
| Windows GUI（系统托盘）| `gopay-X.X.X-windows-amd64.exe` |
| macOS Intel | `gopay-X.X.X-macos-amd64` |
| macOS Apple Silicon | `gopay-X.X.X-macos-arm64` |

> 其中 `X.X.X` 为版本号，如 `1.2.0`。

### 2. 启动

下载后直接运行即可（详见下方各平台说明）。

---

## Linux 部署

### CLI 版本（推荐服务器部署）

```bash
# 1. 下载（替换版本号）
wget https://github.com/your-repo/gopay/releases/download/vX.X.X/gopay-X.X.X-linux-amd64

# 2. 赋予执行权限
chmod +x gopay-*-linux-amd64

# 3. 启动
./gopay-X.X.X-linux-amd64 -host 0.0.0.0 -port 8080
```

启动后访问 `http://你的IP:8080` 进入系统。

### GUI 版本（带系统托盘）

```bash
# 1. 下载
wget https://github.com/your-repo/gopay/releases/download/vX.X.X/gopay-X.X.X-linux-gui-amd64

# 2. 赋予执行权限
chmod +x gopay-*-linux-gui-amd64

# 3. 启动（会显示系统托盘图标）
./gopay-X.X.X-linux-gui-amd64
```

> **注意**: GUI 版本需要桌面环境（X11/Wayland）支持。

### 配置为系统服务（可选）

使用 systemd 管理 GoPay 后台运行：

```bash
# 1. 移动二进制文件
sudo mv gopay-*-linux-amd64 /usr/local/bin/gopay

# 2. 创建 systemd 服务文件
sudo tee /etc/systemd/system/gopay.service > /dev/null << 'EOF'
[Unit]
Description=GoPay Payment Gateway
After=network.target

[Service]
Type=simple
User=gopay
WorkingDirectory=/opt/gopay
ExecStart=/usr/local/bin/gopay -host 0.0.0.0 -port 8080
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

# 3. 创建运行用户和数据目录
sudo useradd -r -s /bin/false gopay
sudo mkdir -p /opt/gopay
sudo chown gopay:gopay /opt/gopay

# 4. 启动服务
sudo systemctl daemon-reload
sudo systemctl enable gopay
sudo systemctl start gopay

# 5. 查看状态
sudo systemctl status gopay
```

### 防火墙配置

```bash
# Ubuntu / Debian
sudo ufw allow 8080/tcp

# CentOS / RHEL
sudo firewall-cmd --permanent --add-port=8080/tcp
sudo firewall-cmd --reload
```

---

## Windows 部署

### GUI 版本

1. 从 [Releases](../../releases) 下载 `gopay-X.X.X-windows-amd64.exe`
2. 将文件放到目标目录（如 `C:\GoPay\`）
3. **双击运行** 或在命令行中启动：

```cmd
gopay-X.X.X-windows-amd64.exe -host 0.0.0.0 -port 8080
```

启动后系统托盘会出现 GoPay 图标，浏览器自动打开管理页面。

### Windows 防火墙

首次启动时 Windows 可能弹出防火墙提示，请选择 **允许访问**。

如需手动放行：

1. 控制面板 → Windows Defender 防火墙 → 高级设置
2. 入站规则 → 新建规则 → 端口 → TCP 8080 → 允许连接

### 开机自启（可选）

**方式一：任务计划程序**

1. Win+R → 输入 `taskschd.msc` → 回车
2. 创建基本任务 → 名称输入 `GoPay`
3. 触发器选择"计算机启动时"
4. 操作选择"启动程序"，浏览选择 `gopay-X.X.X-windows-amd64.exe`
5. 完成

**方式二：启动文件夹**

按 Win+R → 输入 `shell:startup` → 将 GoPay 的快捷方式放入打开的文件夹。

---

## macOS 部署

### 确定芯片类型

- **Intel Mac** (2020 年之前)：下载 `gopay-X.X.X-macos-amd64`
- **Apple Silicon** (M1/M2/M3/M4)：下载 `gopay-X.X.X-macos-arm64`

> 不确定？点击左上角  → 关于本机 → 查看"芯片"信息。

### 安装与启动

```bash
# 1. 下载（替换版本号和架构）
# Apple Silicon:
curl -LO https://github.com/your-repo/gopay/releases/download/vX.X.X/gopay-X.X.X-macos-arm64
# Intel:
curl -LO https://github.com/your-repo/gopay/releases/download/vX.X.X/gopay-X.X.X-macos-amd64

# 2. 赋予执行权限
chmod +x gopay-*-macos-*

# 3. 启动
./gopay-X.X.X-macos-arm64
```

### macOS 安全提示处理

首次运行可能提示"无法验证开发者"：

1. **方式一**：右键点击文件 → 选择"打开" → 在弹窗中点击"打开"
2. **方式二**：系统设置 → 隐私与安全性 → 在底部找到被阻止的应用 → 点击"仍要打开"
3. **方式三（命令行）**：
   ```bash
   xattr -cr gopay-*-macos-*
   ```

---

## 启动参数说明

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `-db` | 数据库文件路径 | 按平台自动选择（见下表） |
| `-host` | 监听 IP 地址 | `0.0.0.0`（所有网卡） |
| `-port` | 监听端口 | `8080` |
| `-migrate` | 执行数据库迁移（版本升级时使用） | `false` |

**示例**：

```bash
# 指定端口和数据库路径
./gopay -host 127.0.0.1 -port 3000 -db /data/gopay.db

# 升级时执行数据库迁移
./gopay -migrate
```

## 默认数据库路径

如不指定 `-db` 参数，数据库文件按以下路径自动创建：

| 平台 | 路径 |
|------|------|
| Windows | `%APPDATA%\gopay\gopay.db` |
| macOS | `~/Library/Application Support/gopay/gopay.db` |
| Linux | `~/.gopay/gopay.db` |

---

## 升级指南

### 标准升级流程

1. **停止服务**
   ```bash
   # systemd 服务
   sudo systemctl stop gopay

   # 或手动停止正在运行的进程
   ```

2. **备份数据库**
   ```bash
   cp ~/.gopay/gopay.db ~/.gopay/gopay.db.bak
   ```

3. **下载新版本** — 从 [Releases](../../releases) 下载最新版本

4. **替换二进制文件**
   ```bash
   # 替换旧版本
   mv gopay-old-linux-amd64 gopay-old-linux-amd64.bak
   mv gopay-new-linux-amd64 gopay
   chmod +x gopay
   ```

5. **启动新版本**（如有数据库变更需加 `-migrate`）
   ```bash
   # 先执行数据库迁移
   ./gopay -migrate

   # 正常启动
   ./gopay -host 0.0.0.0 -port 8080
   ```

6. **验证** — 访问 `http://你的IP:8080` 确认系统正常运行

> **注意**: `-migrate` 参数仅在版本升级说明中有提及数据库变更时使用。

---

## 常见问题排查

### 端口被占用

**现象**: 启动时提示 `bind: address already in use`

**解决**:
```bash
# 查看占用端口的进程
# Linux/macOS:
lsof -i :8080

# Windows:
netstat -ano | findstr 8080

# 解决方式一：更换端口
./gopay -port 8081

# 解决方式二：结束占用进程
kill <PID>
```

### 权限不足

**现象**: 启动时提示 `Permission denied`

**解决**:
```bash
chmod +x gopay-*
```

### 数据库锁定

**现象**: 提示 `database is locked`

**解决**:
1. 确认没有其他 GoPay 实例在运行
2. 检查数据库文件权限：`ls -la ~/.gopay/gopay.db`
3. 如需修复：备份后删除 `.db-journal` 和 `.db-wal` 文件

### 前端页面空白

**现象**: 访问页面显示空白

**解决**:
1. 确认启动日志中没有前端资源相关错误
2. 检查浏览器控制台（F12）是否有错误
3. 确认使用的是包含前端构建的完整 Release 版本

### 无法从外网访问

**排查步骤**:
1. 确认使用 `-host 0.0.0.0` 启动（而非 `127.0.0.1`）
2. 检查防火墙是否放行了端口
3. 检查云服务器安全组是否放行了端口
4. 确认服务器 IP 地址正确

### Windows 被杀毒软件拦截

**解决**:
1. 将 GoPay 添加到杀毒软件白名单/排除列表
2. 或临时关闭杀毒软件的实时防护

### macOS 提示"已损坏无法打开"

**解决**:
```bash
xattr -cr gopay-*-macos-*
```
