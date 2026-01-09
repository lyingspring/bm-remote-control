# Linux 构建指南

由于 Wails 应用使用 CGO（C 语言接口），无法从 macOS 直接交叉编译到 Linux。您需要在 Linux 机器上或在 Linux 环境中进行构建。

## ⚠️ 重要提示

**Wails 框架限制**: 从 macOS 无法直接交叉编译到 Linux。您必须使用以下方法之一：
- 在 Linux 机器上构建
- 使用 Linux 虚拟机
- 使用 Docker 容器
- 使用云服务器

---

## 方法 1: 在 Linux AMD64 机器上构建（最推荐）

### 适用场景
- 目标平台是大多数 PC 和服务器（Intel/AMD 处理器）

### 步骤

#### 1. 将项目传输到 Linux 机器

```bash
# 方法 A: 使用 scp（如果 Linux 机器可通过 SSH 访问）
scp -r bm-remote-control/ user@linux-server:/home/user/

# 方法 B: 使用 rsync（更高效，支持断点续传）
rsync -avz --progress bm-remote-control/ user@linux-server:/home/user/bm-remote-control/

# 方法 C: 使用 Git（如果项目在仓库中）
git clone <your-repo-url>
cd bm-remote-control
```

#### 2. SSH 登录到 Linux 机器

```bash
ssh user@linux-server
cd bm-remote-control
```

#### 3. 安装构建依赖

**Ubuntu/Debian:**
```bash
sudo apt update
sudo apt install -y golang-go nodejs npm libwebkit2gtk-4.0-dev build-essential

# 验证安装
go version
node --version
npm --version
```

**Fedora/RHEL:**
```bash
sudo dnf install -y golang nodejs npm webkit2gtk3-devel gcc make
```

**Arch Linux:**
```bash
sudo pacman -S go nodejs npm webkit2gtk gcc make
```

#### 4. 安装 Wails CLI

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# 确保 Go bin 目录在 PATH 中
export PATH=$PATH:$(go env GOPATH)/bin
echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.bashrc

# 验证安装
wails version
```

#### 5. 运行构建脚本

```bash
chmod +x build-linux-amd64.sh
./build-linux-amd64.sh
```

或手动构建：

```bash
# 安装前端依赖
cd frontend
npm install
cd ..

# 构建
wails build -platform linux/amd64
```

#### 6. 运行应用

```bash
./build/bm-remote-control
```

---

## 方法 2: 在 Linux ARM64 机器上构建

### 适用场景
- 树莓派 4（Raspberry Pi 4）
- NVIDIA Jetson 系列
- Apple Silicon Mac 上的 Linux ARM64 虚拟机
- 云 ARM64 服务器（AWS Graviton、Azure ARM 等）

### 步骤

与 AMD64 方法相同，但使用以下命令：

```bash
chmod +x build-linux-arm64.sh
./build-linux-arm64.sh
```

或手动构建：

```bash
wails build -platform linux/arm64
```

---

## 方法 3: 使用 Docker 构建（高级）

### 使用 Docker 容器构建 AMD64 版本

#### 3.1 安装 Docker

```bash
# macOS
brew install --cask docker

# Linux
curl -fsSL https://get.docker.com | sh
sudo usermod -aG docker $USER
```

#### 3.2 使用 Docker 构建脚本

创建 `docker-build.sh`:

```bash
#!/bin/bash

docker run --rm -v "$(pwd)":/app -w /app \
  -e GOOS=linux \
  -e GOARCH=amd64 \
  node:18-bullseye \
  bash -c "
    apt-get update &&
    apt-get install -y golang-go libwebkit2gtk-4.0-dev build-essential &&
    go install github.com/wailsapp/wails/v2/cmd/wails@latest &&
    cd frontend && npm install && cd .. &&
    export PATH=\$PATH:\$(go env GOPATH)/bin &&
    wails build -platform linux/amd64
  "
```

#### 3.3 运行 Docker 构建

```bash
chmod +x docker-build.sh
./docker-build.sh
```

### 使用现有的 Dockerfile

项目已包含 `Dockerfile.build`:

```bash
# 构建 Docker 镜像
docker build -f Dockerfile.build -t bm-remote-control-builder .

# 运行构建
docker run --rm -v "$(pwd)":/build -w /build bm-remote-control-builder
```

---

## 方法 4: 使用虚拟机

### 在 macOS 上使用虚拟机

#### 选项 A: UTM（免费，推荐）

1. **下载并安装 UTM**
   - 访问: https://utm.app/
   - 下载并安装 UTM

2. **创建 Linux 虚拟机**
   - 下载 Ubuntu Server ISO: https://ubuntu.com/download/server
   - 在 UTM 中创建新的虚拟机
   - 选择 "Virtualize" 模式
   - 分配至少 2GB RAM、20GB 磁盘

3. **在虚拟机中构建**
   - 启动虚拟机并完成 Ubuntu 安装
   - 按照方法 1 的步骤在虚拟机中构建

#### 选项 B: VMware Fusion / Parallels Desktop

1. 创建 Linux 虚拟机
2. 按照方法 1 的步骤构建

---

## 方法 5: 使用云服务器

### 推荐的云服务提供商

**免费/试用选项:**
- **Google Cloud**: 免费试用包含 e2-medium 实例
- **AWS**: 免费套餐包含 12 个月 t2.micro
- **Oracle Cloud**: 永久免费 ARM64 实例

### 步骤

1. **创建云服务器实例**
   - 选择 Ubuntu 22.04 LTS
   - 架构选择 AMD64 或 ARM64
   - 最小配置: 2 vCPU, 4GB RAM

2. **连接到服务器**
   ```bash
   ssh -i your-key.pem ubuntu@your-server-ip
   ```

3. **上传项目**
   ```bash
   # 在本地执行
   scp -i your-key.pem -r bm-remote-control/ ubuntu@your-server-ip:~/
   ```

4. **在服务器上构建**
   ```bash
   ssh -i your-key.pem ubuntu@your-server-ip
   cd bm-remote-control
   ./build-linux-amd64.sh  # 或 build-linux-arm64.sh
   ```

5. **下载构建产物**
   ```bash
   # 在本地执行
   scp -i your-key.pem ubuntu@your-server-ip:~/bm-remote-control/build/bm-remote-control ./
   ```

---

## 快速开始：一键构建命令

如果您已经在 Linux 机器上：

```bash
# AMD64
git clone <your-repo-url> && cd bm-remote-control && chmod +x build-linux-amd64.sh && ./build-linux-amd64.sh

# ARM64
git clone <your-repo-url> && cd bm-remote-control && chmod +x build-linux-arm64.sh && ./build-linux-arm64.sh
```

---

## 构建产物

构建成功后，可执行文件位于：
```
build/bm-remote-control
```

### 验证构建

```bash
# 检查文件类型
file build/bm-remote-control
# 输出应该包含: ELF 64-bit LSB executable...

# 检查依赖
ldd build/bm-remote-control
# 列出动态链接库依赖

# 测试运行
./build/bm-remote-control
```

### 打包分发

```bash
# 创建发布包
tar -czf bm-remote-control-linux-amd64.tar.gz -C build bm-remote-control

# 或使用 zip
zip -r bm-remote-control-linux-amd64.zip build/bm-remote-control
```

---

## 常见问题

### Q: 为什么不能在 macOS 上直接交叉编译到 Linux？
A: Wails 使用 CGO 来调用系统的 WebView 库。不同操作系统的 WebView 实现不同，且 CGO 需要链接到目标平台的系统库，因此交叉编译受到限制。

### Q: 构建时出现 "webview.h not found" 错误
A: 需要安装 webkit2gtk 开发库：
```bash
sudo apt install libwebkit2gtk-4.0-dev
```

### Q: 构建时出现 "cc: command not found" 错误
A: 需要安装构建工具：
```bash
sudo apt install build-essential
```

### Q: 如何为不同的 Linux 发行版构建？
A: 构建出的二进制文件应该是通用的，但如果需要特定发行版的优化，可以在相应的发行版中构建。

### Q: 构建的应用在其他 Linux 机器上无法运行？
A: 可能的原因：
- 目标机器的 glibc 版本太旧
- 缺少 WebView 运行时库
- 解决方案：在较旧的 Linux 发行版（如 Debian 10）上构建以获得更好的兼容性

### Q: 如何创建可安装的 .deb 或 .rpm 包？
A: 可以使用工具如 `fpm` 或 `nfpm`：
```bash
# 安装 nfpm
go install github.com/goreleaser/nfpm/v2/cmd/nfpm@latest

# 创建 .deb 包
nfpm package -p deb
```

---

## 系统要求

### 构建环境要求
- **操作系统**: Linux（推荐 Ubuntu 20.04+ / Debian 11+）
- **Go**: 1.18 或更高版本
- **Node.js**: 16 或更高版本
- **npm**: 7 或更高版本
- **内存**: 至少 2GB RAM（推荐 4GB）
- **磁盘空间**: 至少 1GB 可用空间

### 运行时要求
构建出的应用需要以下运行时库：
- `libwebkit2gtk-4.0-37` 或更高版本
- `libgtk-3-0`
- `glibc` 2.17 或更高版本

Ubuntu/Debian 安装命令：
```bash
sudo apt install libwebkit2gtk-4.0-37 libgtk-3-0
```

---

## 相关资源

- [Wails 官方文档](https://wails.io/docs/introduction)
- [Wails GitHub Issues - Cross Compilation](https://github.com/wailsapp/wails/issues?q=is%3Aissue+cross+compile)
- [Go Cross Compilation](https://go.dev/doc/install/source#environment)
- [Docker Multi-platform Builds](https://docs.docker.com/build/building/multi-platform/)

---

## 技术支持

如果遇到问题：
1. 检查 `wails doctor` 输出
2. 查看 Wails 日志: `wails build -v`
3. 在 [Wails GitHub](https://github.com/wailsapp/wails/issues) 搜索类似问题
4. 查阅项目 CLAUDE.md 中的开发注意事项
