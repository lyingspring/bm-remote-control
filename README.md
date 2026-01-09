# 会声会纪服务端控制面板

一个基于 Wails 框架开发的跨平台 SSH 远程服务管理桌面应用程序，专为会声会纪项目设计，支持通过 SSH 协议远程管理服务、控制系统电源，并实时显示远程主机系统信息。

## 功能特性

- **SSH 远程连接** - 支持密码和 SSH 密钥两种认证方式
- **服务管理** - 一键重启会声会纪核心服务（语音转写、任务处理、后端管理）
- **远程电源控制** - 支持远程关机和重启操作
- **系统信息监控** - 实时显示远程主机的主机名、操作系统、架构和运行时间
- **连接状态监测** - 每 30 秒自动心跳检测，实时显示连接状态
- **配置持久化** - SSH 配置自动保存到本地 settings.json 文件
- **安全设计** - 移除任意命令执行功能，仅允许管理指定服务

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

## 管理的服务

应用程序专为以下会声会纪服务设计：

1. **语音转写服务** (`voice-transcription-celery.service`)
   - 负责语音转文字处理

2. **任务处理系统** (`ga-tingfeng-api-celery.service`)
   - 异步任务队列处理

3. **后端管理服务** (`ga-tingfeng-api.service`)
   - 核心业务逻辑服务

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
- 桌面应用窗口
- Vite 开发服务器: http://localhost:5173

### 生产构建

```bash
wails build
```

构建产物将根据平台生成在 `build/` 目录下。

## 使用说明

### 首次配置

1. 启动应用后，点击右上角 **"⚙️ 设置"** 按钮
2. 填写 SSH 连接信息：
   - **IP 地址**：远程服务器 IP（例如：192.168.1.100）
   - **端口**：SSH 端口（默认 22）
   - **用户名**：SSH 登录用户名
   - **密码**：SSH 密码（可选，留空则使用密钥认证）
3. 点击 **"保存配置"** 或 **"测试连接"** 验证连接

### 连接状态

应用会自动监控 SSH 连接状态：
- 🟢 **绿色** - 已连接
- 🔴 **红色** - 未连接
- 🟡 **黄色** - 检查中...

每 30 秒自动执行一次心跳检测。

### 服务管理

点击对应服务的 **"↻ 重启服务"** 按钮，确认后即可重启服务：
- 语音转写服务
- 任务处理系统
- 后端管理服务

### 远程电源控制

- **远程关机** - 立即关闭远程主机
- **远程重启** - 立即重启远程主机

⚠️ **注意**：电源控制需要 sudo 权限，请确保远程用户已配置无密码 sudo。

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

5. **权限控制** - 电源控制和服务重启需要 sudo 权限。请确保远程用户已正确配置：
   ```bash
   sudo visudo
   # 添加以下行
   username ALL=(ALL) NOPASSWD: ALL
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

### Q: 服务重启失败？
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

## 前端样式

项目使用 Tailwind CSS 进行样式设计，主要样式特点：

- 深色主题 (gray-900 背景)
- 响应式布局 (grid/flex)
- 颜色编码的输出显示:
  - 绿色 - 成功/正常输出
  - 黄色 - 警告
  - 红色 - 错误
  - 蓝色 - 进行中

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

## 开发历史

- 初始版本基于本地电脑控制功能
- 添加 SSH 远程控制功能
- 移除所有本地操作，专注于远程控制
- 从 JavaScript 生成 HTML 转换为传统 HTML 文件
- 集成 Tailwind CSS 进行样式设计
- 修复 Tailwind CSS v4 兼容性问题，降级到 v3.4
- 移除任意命令执行功能，增强安全性
- 添加会声会纪专属服务管理功能
- 更新为"会声会纪服务端控制面板"

## 许可证

本项目为示例项目，仅供学习和测试使用。

## 贡献

欢迎提交 Issue 和 Pull Request！
