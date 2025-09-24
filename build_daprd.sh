#!/bin/bash

# Dapr程序交叉编译脚本
# 支持Linux和Windows的常用架构

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 打印带颜色的消息
print_info() {
  echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
  echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_error() {
  echo -e "${RED}[ERROR]${NC} $1"
}

print_warning() {
  echo -e "${YELLOW}[WARNING]${NC} $1"
}

# 版本信息配置
VERSION=${VERSION:-"edge"}
REL_VERSION=${REL_VERSION:-$VERSION}
GIT_COMMIT=${GIT_COMMIT:-$(git rev-list -1 HEAD 2>/dev/null || echo "unknown")}
GIT_VERSION=${GIT_VERSION:-$(git describe --always --abbrev=7 --dirty 2>/dev/null || echo "unknown")}

# 显示版本信息
print_info "版本配置信息："
echo "  - 版本号: $REL_VERSION"
echo "  - Git Commit: $GIT_COMMIT"
echo "  - Git Version: $GIT_VERSION"
echo ""

# 编译平台配置
PLATFORMS=(
  "linux:amd64:Linux 64位"
  "linux:arm64:Linux ARM64"
  "windows:amd64:Windows 64位"
)

# 开始编译
print_info "开始Dapr程序交叉编译..."
echo ""

success_count=0
total_count=${#PLATFORMS[@]}

for platform_info in "${PLATFORMS[@]}"; do
  IFS=':' read -ra INFO <<< "$platform_info"
  GOOS=${INFO[0]}
  GOARCH=${INFO[1]}
  DESC=${INFO[2]}
  
  print_info "正在编译 $DESC ($GOOS/$GOARCH)..."
  
  # 执行编译命令（添加版本信息）
  if make build GOOS=$GOOS GOARCH=$GOARCH BINARIES=daprd REL_VERSION=$REL_VERSION GIT_COMMIT=$GIT_COMMIT GIT_VERSION=$GIT_VERSION; then
      print_success "$DESC 编译成功"
      ((success_count++))
  else
      print_error "$DESC 编译失败"
  fi
  
  echo ""
done

# 编译结果统计
echo "=================================="
if [ $success_count -eq $total_count ]; then
  print_success "所有平台编译完成！($success_count/$total_count)"
else
  print_warning "编译完成，成功: $success_count/$total_count"
fi
echo "=================================="

# 检查生成的文件
if [ -d "dist" ]; then
  print_info "生成的文件："
  ls -la dist/
  
  # 显示版本信息验证
  echo ""
  print_info "版本信息验证："
  for platform_info in "${PLATFORMS[@]}"; do
    IFS=':' read -ra INFO <<< "$platform_info"
    GOOS=${INFO[0]}
    GOARCH=${INFO[1]}
    DESC=${INFO[2]}
    
    BINARY_PATH="dist/${GOOS}_${GOARCH}/release/daprd"
    if [ "$GOOS" = "windows" ]; then
      BINARY_PATH="${BINARY_PATH}.exe"
    fi
    
    if [ -f "$BINARY_PATH" ]; then
      echo "  - $DESC: $BINARY_PATH"
      # 可以通过 $BINARY_PATH --version 来验证版本（如果二进制支持的话）
    fi
  done
fi

echo ""
if [ -z "$VERSION" ] || [ -z "$REL_VERSION" ]; then
  print_info "使用说明："
  echo "  - 设置版本: VERSION=v1.12.0 ./build_daprd.sh"
  echo "  - 设置发布版本: REL_VERSION=v1.12.0 ./build_daprd.sh"
  echo "  - 完整示例: VERSION=v1.12.0 REL_VERSION=v1.12.0 ./build_daprd.sh"
fi