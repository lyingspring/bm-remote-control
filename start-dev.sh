#!/bin/bash

# Wails 开发环境启动脚本 (Ubuntu/Debian)
# 使用方法: ./start-dev.sh

# 设置 Go 路径 (如果使用自定义安装)
if [ -d "$HOME/go-install/go/bin" ]; then
    export PATH="$HOME/go-install/go/bin:$PATH"
fi

if [ -d "$HOME/go/bin" ]; then
    export PATH="$HOME/go/bin:$PATH"
fi

echo "========================================="
echo "  Wails 开发服务器启动脚本"
echo "========================================="

# 检查是否安装了必要的依赖
echo "[1/4] 检查系统依赖..."

if ! dpkg -l | grep -q libgtk-3-dev; then
    echo "❌ 缺少 libgtk-3-dev"
    echo "请运行: sudo apt-get install libgtk-3-dev"
    exit 1
fi

if ! dpkg -l | grep -q libwebkit2gtk-4.1-dev; then
    echo "❌ 缺少 libwebkit2gtk-4.1-dev"
    echo "请运行: sudo apt-get install libwebkit2gtk-4.1-dev"
    exit 1
fi

echo "✓ 系统依赖检查通过"

# 检查 Go 环境
echo "[2/4] 检查 Go 环境..."

if ! command -v go &> /dev/null; then
    echo "❌ 未找到 Go"
    echo "请确保 Go 已安装并在 PATH 中"
    exit 1
fi

GO_VERSION=$(go version | awk '{print $3}')
echo "✓ Go 版本: $GO_VERSION"

# 检查 Wails CLI
echo "[3/4] 检查 Wails CLI..."

if ! command -v wails &> /dev/null; then
    echo "❌ 未找到 Wails CLI"
    echo "请运行: go install github.com/wailsapp/wails/v2/cmd/wails@latest"
    exit 1
fi

WAILS_VERSION=$(wails version | head -n 1)
echo "✓ Wails 版本: $WAILS_VERSION"

# 设置环境变量
echo "[4/4] 配置环境..."

# 设置 Go 路径 (如果使用自定义安装)
if [ -d "$HOME/go-install/go/bin" ]; then
    export PATH="$PATH:$HOME/go-install/go/bin"
fi

if [ -d "$HOME/go/bin" ]; then
    export PATH="$PATH:$HOME/go/bin"
fi

# 检测是否需要设置 DISPLAY
if [ -z "$DISPLAY" ]; then
    echo "⚠️  DISPLAY 环境变量未设置"
    echo "如果是远程 SSH 连接,请确保启用了 X11 转发"
    echo "或者直接访问浏览器开发界面: http://localhost:34115"
    echo ""
    read -p "是否继续? (y/n) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
else
    echo "✓ DISPLAY=$DISPLAY"
fi

# 启动 Wails 开发服务器
echo ""
echo "========================================="
echo "  启动 Wails 开发服务器"
echo "========================================="
echo ""
echo "使用 webkit2gtk-4.1 编译标签..."
echo ""
echo "开发服务器地址:"
echo "  - 浏览器开发界面: http://localhost:34115"
echo "  - 前端开发服务器: http://localhost:5173"
echo ""
echo "按 Ctrl+C 停止服务器"
echo ""

# 启动 Wails (使用 webkit2_41 标签以支持 Ubuntu 24.04+)
wails dev -tags webkit2_41
