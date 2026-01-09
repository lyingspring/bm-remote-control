#!/bin/bash

# Linux ARM64 构建脚本
# 此脚本需要在 Linux ARM64 机器上运行（如树莓派 4、云服务器等）

set -e

echo "========================================="
echo "  Linux ARM64 构建脚本"
echo "========================================="
echo ""

# 检查系统架构
ARCH=$(uname -m)
if [ "$ARCH" != "aarch64" ]; then
    echo "警告: 当前系统架构是 $ARCH，不是 aarch64 (ARM64)"
    echo "此脚本需要在 Linux ARM64 机器上运行"
    read -p "是否继续？(y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

# 检查必要的工具
echo "检查构建依赖..."
for cmd in go node npm wails; do
    if ! command -v $cmd &> /dev/null; then
        echo "错误: 未找到 $cmd"
        echo "请先安装必要的依赖:"
        echo "  sudo apt update"
        echo "  sudo apt install -y golang-go nodejs npm libwebkit2gtk-4.0-dev build-essential"
        echo "  go install github.com/wailsapp/wails/v2/cmd/wails@latest"
        exit 1
    fi
done

echo "✓ 所有依赖已就绪"
echo ""

# 显示版本信息
echo "构建环境信息:"
echo "  Go: $(go version)"
echo "  Node: $(node --version)"
echo "  npm: $(npm --version)"
echo "  Wails: $(wails version)"
echo "  系统: $(uname -a)"
echo ""

# 清理旧的构建产物
echo "清理旧构建..."
rm -rf build/
mkdir -p build
echo "✓ 清理完成"
echo ""

# 安装前端依赖
echo "安装前端依赖..."
cd frontend
npm install
cd ..
echo "✓ 前端依赖安装完成"
echo ""

# 开始构建
echo "开始构建 Linux ARM64 版本..."
echo "这可能需要几分钟时间..."
echo ""

wails build -platform linux/arm64 -clean

echo ""
echo "========================================="
echo "  构建完成！"
echo "========================================="
echo ""
echo "构建产物:"
ls -lh build/
echo ""
echo "可执行文件位置: build/bm-remote-control"
echo ""
echo "要运行应用:"
echo "  ./build/bm-remote-control"
echo ""
