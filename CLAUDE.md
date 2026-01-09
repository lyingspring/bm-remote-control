# SSH 远程控制系统

一个基于 Wails 框架开发的跨平台 SSH 远程控制桌面应用程序，支持通过 SSH 协议远程执行命令、控制系统电源，并显示远程主机系统信息。

## 功能特性

- **SSH 远程连接** - 支持密码和 SSH 密钥两种认证方式
- **远程命令执行** - 在远程主机上执行任意 shell 命令并查看输出
- **远程电源控制** - 支持远程关机和重启操作
- **系统信息显示** - 实时显示远程主机的主机名、操作系统、架构和运行时间
- **配置持久化** - SSH 配置自动保存到本地 settings.json 文件

## 技术栈

### 后端
- **Go 1.x** - 主要编程语言
- **Wails v2.10.1** - 桌面应用框架
- **golang.org/x/crypto/ssh** - 纯 Go SSH 客户端实现

### 前端
- **原生 JavaScript (ES Modules)** - 无框架依赖
- **Tailwind CSS v3.4** - 实用优先的 CSS 框架
- **Vite v3.2** - 前端构建工具
- **PostCSS** - CSS 转换工具

## 项目结构

```
bm-remote-control/
├── app.go                    # 后端 Go 应用主文件
├── main.go                   # 应用入口
├── wails.json                # Wails 配置文件
├── settings.json             # SSH 配置文件（运行时生成）
├── frontend/
│   ├── index.html            # 主 HTML 页面
│   ├── package.json          # 前端依赖配置
│   ├── tailwind.config.js    # Tailwind CSS 配置
│   ├── postcss.config.js     # PostCSS 配置
│   ├── vite.config.js        # Vite 配置
│   └── src/
│       ├── main.js           # 前端 JavaScript 逻辑
│       ├── tailwind.css      # Tailwind CSS 入口
│       └── assets/
│           └── images/
│               └── logo-universal.png
└── wailsjs/                  # Wails 自动生成的绑定
    ├── go/
    │   └── main/
    │       └── App.js        # Go 方法绑定
    └── runtime/
        └── ...
```

## 安装和运行

### 前置要求

- Go 1.18 或更高版本
- Node.js 16 或更高版本
- Wails CLI v2.10.1

### 安装步骤

1. 安装 Wails CLI（如果尚未安装）:
```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

2. 安装前端依赖:
```bash
cd frontend
npm install
```

### 开发模式

在项目根目录运行:
```bash
wails dev
```

应用将自动编译并启动开发服务器，默认访问地址：
- 浏览器开发界面: http://localhost:34115
- Vite 开发服务器: http://localhost:5173

### 生产构建

```bash
wails build
```

构建产物将根据平台生成在 `build/` 目录下。

## API 接口文档

### 前端调用的 Go 方法

#### 1. SaveSSHConfig
保存 SSH 连接配置到 settings.json

**参数:**
- `host` (string) - SSH 服务器 IP 地址
- `port` (string) - SSH 端口，默认 22
- `username` (string) - SSH 用户名
- `password` (string) - SSH 密码（可选，留空则使用密钥认证）

**返回:** 成功消息字符串

**示例:**
```javascript
SaveSSHConfig("192.168.1.100", "22", "user", "password")
```

#### 2. LoadSSHConfig
从 settings.json 加载 SSH 配置

**返回:** 包含配置的 map[string]interface{}
```json
{
  "host": "192.168.1.100",
  "port": "22",
  "username": "user",
  "password": "password"
}
```

#### 3. TestSSHConnection
测试当前 SSH 配置的连接状态

**返回:** 成功消息字符串

**示例:**
```javascript
TestSSHConnection() // "SSH 连接测试成功！"
```

#### 4. ExecuteSSHCommand
在远程主机上执行命令

**参数:**
- `command` (string) - 要执行的 shell 命令

**返回:** 命令输出字符串

**示例:**
```javascript
ExecuteSSHCommand("ls -la") // 返回文件列表
```

#### 5. GetRemoteSystemInfo
获取远程主机系统信息

**返回:** 包含系统信息的 map[string]interface{}
```json
{
  "hostname": "remote-host",
  "os": "Linux",
  "arch": "x86_64",
  "uptime": "up 3 weeks, 4 days"
}
```

### HTTP API 接口

应用同时在端口 8080 启动 HTTP 服务器，提供以下 REST API：

#### POST /api/shutdown
发送关机命令（本地机器）

**响应:**
```json
{"success": true, "message": "Shutdown command sent"}
```

#### POST /api/restart
发送重启命令（本地机器）

**响应:**
```json
{"success": true, "message": "Restart command sent"}
```

#### POST /api/sleep
发送休眠命令（本地机器）

**响应:**
```json
{"success": true, "message": "Sleep command sent"}
```

#### POST /api/command
在本地机器执行命令

**请求体:** 原始命令字符串

**响应:**
```json
{
  "success": true,
  "output": "命令输出",
  "error": ""
}
```

#### GET /api/status
获取应用状态

**响应:**
```json
{
  "status": "running",
  "os": "darwin",
  "arch": "arm64"
}
```

#### GET /health
健康检查端点

**响应:** `OK`

## SSH 认证方式

应用支持两种 SSH 认证方式：

### 1. 密码认证
在配置界面直接输入 SSH 密码，系统将使用密码进行认证。

### 2. SSH 密钥认证
将密码字段留空，系统将自动尝试使用以下私钥文件（按优先级）：
- `~/.ssh/id_ed25519` (推荐)
- `~/.ssh/id_rsa`
- `~/.ssh/id_ecdsa`

**注意:** 私钥文件不应有密码保护，否则将被跳过。

## 安全注意事项

⚠️ **重要安全提示:**

1. **Host Key Verification** - 当前实现使用 `ssh.InsecureIgnoreHostKey()`，不验证主机密钥。生产环境应使用 `ssh.FixedHostKey()` 进行严格的主机密钥验证。

2. **密码存储** - 密码以明文形式存储在 settings.json 中。生产环境应考虑使用操作系统密钥链或加密存储。

3. **文件权限** - settings.json 文件权限设置为 0600，仅所有者可读写。

4. **网络传输** - SSH 协议本身是加密的，但请确保使用强密码和安全的密钥。

## 前端样式

项目使用 Tailwind CSS 进行样式设计，主要样式特点：

- **离线优先设计** - 不依赖任何外部 CDN 或在线资源
- **深色主题** - 紫罗兰色系 (purple/violet)，科技感界面
- **响应式布局** - grid/flex 布局
- **动画效果** - 扫描线背景、网格动画、脉动状态灯
- **颜色编码的输出显示:**
  - 绿色 - 成功/正常输出
  - 黄色 - 警告
  - 红色 - 错误
  - 蓝色 - 信息提示

### 离线优先原则

⚠️ **重要：本应用为离线桌面应用，严禁使用远程资源！**

1. **字体使用系统字体栈**
   - 等宽字体: `SF Mono, Monaco, Menlo, Consolas, Courier New, monospace`
   - 正文字体: `-apple-system, BlinkMacSystemFont, Segoe UI, Roboto, Helvetica Neue, Arial, sans-serif`
   - 不使用 Google Fonts 或其他在线字体服务

2. **禁止使用外部 CDN**
   - 不引用任何外部 CSS/JS 库
   - 不使用 CDN 加载字体、图标或图片
   - 所有资源必须本地化

3. **图标使用内联 SVG**
   - 所有图标直接嵌入 HTML
   - 不依赖图标字体库（如 Font Awesome）

### 关键 CSS 类

```html
<!-- 卡片容器 -->
<div class="bg-gray-800 rounded-lg p-6 shadow-lg">

<!-- 输入框 -->
<input class="w-full px-4 py-2 bg-gray-700 text-white rounded-lg border border-gray-600 focus:border-blue-500 focus:outline-none">

<!-- 按钮 -->
<button class="bg-green-600 hover:bg-green-700 text-white font-bold py-2 px-4 rounded-lg transition">

<!-- 输出区域 -->
<pre class="bg-gray-900 text-green-400 p-4 rounded-lg border-l-4 border-green-600 font-mono text-sm">
```

## 开发注意事项

### Wails 框架限制

1. **原生浏览器对话框不可用**
   - 在 Wails 中，原生的 `confirm()`、`alert()` 和 `prompt()` 对话框无法正常显示
   - **解决方案:** 使用自定义的 HTML/CSS 模态框
   - 项目已实现自定义确认对话框 (`#confirmDialog`)，位于 [index.html:55-73](frontend/index.html#L55-L73)
   - 使用 Promise 模式实现异步确认流程

2. **自定义确认对话框使用示例**
```javascript
// 定义确认对话框函数
function showConfirm(title, message) {
    return new Promise((resolve) => {
        document.getElementById('confirmTitle').textContent = title;
        document.getElementById('confirmMessage').textContent = message;
        confirmCallback = resolve;
        confirmDialog.classList.remove('hidden');
        document.body.style.overflow = 'hidden';
    });
}

// 使用 async/await 调用
window.remoteShutdown = async function() {
    const confirmed = await showConfirm('远程关机', '确定要关闭远程主机吗？');
    if (!confirmed) return;
    // 执行关机操作...
};
```

### Sudo 密码自动输入

1. **问题:** 远程执行 sudo 命令时需要交互式输入密码
2. **解决方案:** 使用 `sudo -S` 参数从 stdin 读取密码
3. **实现位置:** [app.go:456-500](app.go#L456-L500)

**关键代码:**
```go
if strings.HasPrefix(command, "sudo ") && config.Password != "" {
    // 添加 -S 参数让 sudo 从 stdin 读取密码
    sudoCommand := strings.Replace(command, "sudo ", "sudo -S ", 1)

    // 创建 stdin 管道
    stdin, err := session.StdinPipe()

    // 收集输出
    var outputBuf bytes.Buffer
    session.Stdout = &outputBuf
    session.Stderr = &outputBuf

    // 启动命令
    session.Start(sudoCommand)

    // 通过 goroutine 立即输入密码
    go func() {
        defer stdin.Close()
        fmt.Fprintf(stdin, "%s\n", config.Password)
    }()

    // 等待命令完成
    session.Wait()
}
```

4. **注意事项:**
   - 需要导入 `bytes` 包
   - 密码输入必须在 goroutine 中异步执行
   - 使用 `-S` 参数而不是 `-p` 参数
   - 需要处理密码错误的情况，尝试回退到无密码方式

### 连接状态监测

1. **心跳检测机制:**
   - 每 30 秒自动检查一次 SSH 连接状态
   - 使用轻量级 `echo "heartbeat"` 命令进行检测
   - 实现位置: [main.js:133-164](frontend/src/main.js#L133-L164)

2. **状态指示器:**
   - 🟢 绿色 + "已连接" - SSH 连接正常
   - 🔴 红色 + "未连接" - SSH 连接失败
   - 🟡 黄色 + "检查中..." - 正在检测连接
   - UI 位置: [index.html:167-170](frontend/index.html#L167-L170)

3. **启动方式:**
```javascript
// 页面加载时自动启动
window.onload = function() {
    loadSSHConfig();
    loadRemoteSystemInfo();
    startHeartbeat(); // 启动心跳监测
};
```

### Toast 通知系统

1. **用于非侵入式用户反馈**
2. **支持四种类型:** success (✓), error (✕), info (ℹ), warning (⚠)
3. **自动 3 秒后淡出移除**
4. **实现位置:** [main.js:48-84](frontend/src/main.js#L48-L84)

### Z-Index 层级管理

由于页面有多个浮层组件，需要严格管理 z-index：

- `z-[60]` - Toast 通知容器
- `z-[70]` - 确认对话框
- `z-50` - SSH 配置模态框
- `z-40` - 设置按钮

**重要:** 确保 z-index 层级正确，避免模态框被其他元素遮挡。

### 远程电源控制命令优先级

由于不同 Linux 发行版的命令可能不同，系统使用命令重试机制：

1. **优先使用 systemd 命令** (较新的发行版)
2. **其次使用传统命令** (兼容旧系统)
3. **优先尝试带 sudo 的命令** (因为会自动输入密码)
4. **实现位置:** [main.js:247-268](frontend/src/main.js#L247-L268)

```javascript
const shutdownCommands = [
    'sudo systemctl poweroff',  // 优先：systemd + sudo
    'sudo shutdown -h now',
    'sudo poweroff',
    'systemctl poweroff',       // 回退：无 sudo
    'shutdown -h now',
    'poweroff'
];
```

## 常见问题

### Q: Tailwind CSS 样式不加载？
A: 确保 Tailwind CSS v3.4+ 正确安装，检查 postcss.config.js 和 tailwind.config.js 配置是否正确。如果之前使用过 v4，需要完全卸载并清理 node_modules。

```bash
npm uninstall tailwindcss @tailwindcss/postcss
npm install -D tailwindcss@^3.4.1 autoprefixer@^10.4.16 postcss@^8.4.31
rm -rf node_modules/.vite dist
```

### Q: SSH 连接失败？
A: 检查以下几点：
- IP 地址和端口是否正确
- 用户名和密码是否正确
- 远程主机 SSH 服务是否运行
- 网络连接是否正常
- 如果使用密钥认证，检查密钥文件是否存在且权限正确

### Q: 点击远程重启/关机按钮没有反应？
A: 检查以下几点：
- 确认自定义对话框是否正常显示（Wails 不支持原生 confirm）
- 查看浏览器控制台是否有 JavaScript 错误
- 确认 SSH 连接状态指示器是否显示"已连接"
- 检查 sudo 命令是否需要密码，是否正确配置了自动输入

### Q: Sudo 命令执行失败？
A:
1. 确认已配置 SSH 密码（用于 sudo 自动输入）
2. 确认远程用户已配置 sudo 权限
3. 或者配置无密码 sudo: 在远程主机执行 `sudo visudo` 添加 `username ALL=(ALL) NOPASSWD: ALL`

### Q: 如何启用 SSH agent 认证？
A: 当前版本的 SSH agent 认证功能是占位实现。如需启用，需要修改 `getSSHAgentAuth()` 函数，使用 `golang.org/x/crypto/ssh/agent` 包实现完整的 agent 支持。

### Q: 支持的操作系统？
A:
- **开发平台:** macOS, Linux, Windows
- **远程主机:** 任何支持 SSH 的 Unix-like 系统 (Linux, macOS, BSD)

## 开发历史

- 初始版本基于本地电脑控制功能
- 添加 SSH 远程控制功能
- 移除所有本地操作，专注于远程控制
- 从 JavaScript 生成 HTML 转换为传统 HTML 文件
- 集成 Tailwind CSS 进行样式设计
- 修复 Tailwind CSS v4 兼容性问题，降级到 v3.4

## 许可证

本项目为示例项目，仅供学习和测试使用。

## 贡献

欢迎提交 Issue 和 Pull Request！
