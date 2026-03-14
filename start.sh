#!/bin/bash

##################################################################################
#                                                                                #
#                    TGBot Admin 一键部署脚本                                     #
#                                                                                #
#  支持:                                                                         #
#    - 从零环境部署 (自动安装 Docker)                                             #
#    - Docker Hub 拉取部署 (推荐)                                                 #
#    - 源码构建部署                                                               #
#    - 多种操作系统: Ubuntu, Debian, CentOS, macOS                               #
#                                                                                #
##################################################################################

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Version
VERSION="1.1.0"
DOCKER_HUB_IMAGE="nodesire7/tgbot-admin"
GITHUB_REPO="https://github.com/nodesire7/TGBot_Admin.git"

# Project directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Detect OS
detect_os() {
    if [[ "$OSTYPE" == "darwin"* ]]; then
        echo "macos"
    elif [ -f /etc/os-release ]; then
        . /etc/os-release
        echo "$ID"
    else
        echo "unknown"
    fi
}

OS=$(detect_os)

# Print banner
print_banner() {
    echo -e "${CYAN}"
    echo "╔════════════════════════════════════════════════════════════╗"
    echo "║                                                            ║"
    echo "║              TGBot Admin - 一键部署脚本                    ║"
    echo "║                                                            ║"
    echo "║       Go API + Python Bot + Tailwind UI + Docker          ║"
    echo "║                                                            ║"
    echo "╚════════════════════════════════════════════════════════════╝"
    echo -e "${NC}"
    echo -e "版本: ${GREEN}$VERSION${NC}"
    echo -e "系统: ${GREEN}$OS${NC}"
    echo ""
}

# Print error and exit
error_exit() {
    echo -e "${RED}错误: $1${NC}"
    exit 1
}

# Check if running as root
check_root() {
    if [[ $EUID -ne 0 ]] && [[ "$OS" != "macos" ]]; then
        echo -e "${YELLOW}建议使用 sudo 运行此脚本${NC}"
        echo ""
    fi
}

# Install Docker on Ubuntu/Debian
install_docker_debian() {
    echo -e "${YELLOW}正在安装 Docker...${NC}"

    # Update packages
    apt-get update

    # Install dependencies
    apt-get install -y \
        ca-certificates \
        curl \
        gnupg \
        lsb-release

    # Add Docker GPG key
    mkdir -p /etc/apt/keyrings
    curl -fsSL https://download.docker.com/linux/$ID/gpg | gpg --dearmor -o /etc/apt/keyrings/docker.gpg

    # Add Docker repository
    echo \
        "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/$ID \
        $(lsb_release -cs) stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null

    # Install Docker
    apt-get update
    apt-get install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin

    # Start Docker
    systemctl start docker
    systemctl enable docker

    # Add current user to docker group
    local CURRENT_USER=${SUDO_USER:-$USER}
    usermod -aG docker "$CURRENT_USER" 2>/dev/null || true

    echo -e "${GREEN}✓ Docker 安装完成${NC}"
}

# Install Docker on CentOS/RHEL
install_docker_centos() {
    echo -e "${YELLOW}正在安装 Docker...${NC}"

    # Install dependencies
    yum install -y yum-utils

    # Add Docker repository
    yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo

    # Install Docker
    yum install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin

    # Start Docker
    systemctl start docker
    systemctl enable docker

    # Add current user to docker group
    local CURRENT_USER=${SUDO_USER:-$USER}
    usermod -aG docker "$CURRENT_USER" 2>/dev/null || true

    echo -e "${GREEN}✓ Docker 安装完成${NC}"
}

# Install Docker on macOS
install_docker_macos() {
    echo -e "${YELLOW}正在检查 Docker...${NC}"

    if command -v brew &> /dev/null; then
        echo -e "${YELLOW}使用 Homebrew 安装 Docker...${NC}"
        brew install --cask docker
    else
        echo -e "${YELLOW}请安装 Homebrew 或手动下载 Docker Desktop:${NC}"
        echo "https://www.docker.com/products/docker-desktop"
        read -p "已安装 Docker Desktop? 按 Enter 继续..."
    fi

    # Wait for Docker to start
    echo -e "${YELLOW}等待 Docker 启动...${NC}"
    sleep 10

    echo -e "${GREEN}✓ Docker 就绪${NC}"
}

# Install Docker
install_docker() {
    echo -e "${YELLOW}检测到 Docker 未安装${NC}"
    echo ""

    read -p "是否自动安装 Docker? (Y/n): " install_choice
    install_choice=${install_choice:-Y}

    if [[ "$install_choice" =~ ^[Yy]$ ]]; then
        case $OS in
            ubuntu|debian|linuxmint|pop)
                install_docker_debian
                ;;
            centos|rhel|fedora|rocky|almalinux)
                install_docker_centos
                ;;
            macos)
                install_docker_macos
                ;;
            *)
                error_exit "不支持的系统: $OS，请手动安装 Docker"
                ;;
        esac
    else
        error_exit "需要 Docker 才能继续部署"
    fi
}

# Check Docker
check_docker() {
    echo -e "${YELLOW}检查 Docker...${NC}"

    if ! command -v docker &> /dev/null; then
        install_docker
    fi

    # Check if Docker is running
    if ! docker info &> /dev/null; then
        echo -e "${YELLOW}Docker 未运行，正在启动...${NC}"

        if [[ "$OS" == "macos" ]]; then
            open -a Docker 2>/dev/null || true
            sleep 15
        else
            systemctl start docker 2>/dev/null || service docker start 2>/dev/null || true
        fi

        # Wait for Docker
        for i in {1..30}; do
            if docker info &> /dev/null; then
                break
            fi
            sleep 1
        done
    fi

    if ! docker info &> /dev/null; then
        error_exit "Docker 启动失败"
    fi

    echo -e "${GREEN}✓ Docker 运行正常${NC}"
}

# Check Docker Compose
check_docker_compose() {
    echo -e "${YELLOW}检查 Docker Compose...${NC}"

    if docker compose version &> /dev/null; then
        echo -e "${GREEN}✓ Docker Compose 已安装${NC}"
    elif command -v docker-compose &> /dev/null; then
        echo -e "${GREEN}✓ Docker Compose 已安装 (独立版本)${NC}"
    else
        echo -e "${YELLOW}Docker Compose 未安装${NC}"

        if [[ "$OS" != "macos" ]]; then
            # Install docker-compose as fallback
            curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" \
                -o /usr/local/bin/docker-compose
            chmod +x /usr/local/bin/docker-compose
            echo -e "${GREEN}✓ Docker Compose 安装完成${NC}"
        else
            error_exit "请确保 Docker Desktop 已正确安装"
        fi
    fi
}

# Docker compose command wrapper
docker_compose() {
    if docker compose version &> /dev/null; then
        docker compose "$@"
    else
        docker-compose "$@"
    fi
}

# Check .env file
check_env() {
    if [ ! -f "$SCRIPT_DIR/.env" ]; then
        echo -e "${YELLOW}创建 .env 配置文件...${NC}"
        cp "$SCRIPT_DIR/.env.example" "$SCRIPT_DIR/.env"

        # Generate random secrets
        JWT_SECRET=$(openssl rand -hex 32 2>/dev/null || cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 64 | head -n 1)
        REDIS_PASSWORD=$(openssl rand -hex 16 2>/dev/null || cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 32 | head -n 1)

        # Update .env
        if [[ "$OSTYPE" == "darwin"* ]]; then
            sed -i '' "s/your_super_secret_jwt_key_change_this_in_production/${JWT_SECRET}/g" "$SCRIPT_DIR/.env"
            sed -i '' "s/your_redis_password/${REDIS_PASSWORD}/g" "$SCRIPT_DIR/.env"
        else
            sed -i "s/your_super_secret_jwt_key_change_this_in_production/${JWT_SECRET}/g" "$SCRIPT_DIR/.env"
            sed -i "s/your_redis_password/${REDIS_PASSWORD}/g" "$SCRIPT_DIR/.env"
        fi

        echo -e "${GREEN}✓ .env 已创建${NC}"
    fi
}

# Check BOT_TOKEN
check_bot_token() {
    if grep -q "your_bot_token_here" "$SCRIPT_DIR/.env" 2>/dev/null; then
        echo ""
        echo -e "${RED}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
        echo -e "${RED}               BOT_TOKEN 未配置！${NC}"
        echo -e "${RED}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
        echo ""
        echo -e "${CYAN}获取 Bot Token 步骤：${NC}"
        echo "  1. 在 Telegram 搜索 @BotFather"
        echo "  2. 发送 /newbot 创建新 Bot"
        echo "  3. 按提示设置 Bot 名称"
        echo "  4. 复制获得的 Token"
        echo ""
        echo -e "${CYAN}配置方法：${NC}"
        echo "  编辑 $SCRIPT_DIR/.env"
        echo "  设置 BOT_TOKEN=你的token"
        echo ""

        read -p "请输入 Bot Token: " bot_token
        if [[ -n "$bot_token" ]]; then
            if [[ "$OSTYPE" == "darwin"* ]]; then
                sed -i '' "s/your_bot_token_here/${bot_token}/g" "$SCRIPT_DIR/.env"
            else
                sed -i "s/your_bot_token_here/${bot_token}/g" "$SCRIPT_DIR/.env"
            fi
            echo -e "${GREEN}✓ Bot Token 已保存${NC}"
        else
            echo -e "${YELLOW}跳过配置，稍后请手动设置${NC}"
        fi
    fi
}

# Create docker-compose.hub.yml for Docker Hub deployment
create_hub_compose() {
    cat > "$SCRIPT_DIR/docker-compose.hub.yml" << 'EOF'
version: '3.8'

services:
  tgbot-admin:
    image: nodesire7/tgbot-admin:latest
    container_name: tgbot_admin
    restart: unless-stopped
    ports:
      - "${API_PORT:-8000}:8000"
    environment:
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_USER=${DB_USER:-tgbot}
      - DB_PASSWORD=${DB_PASSWORD:-tgbot123}
      - DB_NAME=${DB_NAME:-tgbot}
      - REDIS_HOST=redis
      - REDIS_PORT=6379
      - REDIS_PASSWORD=${REDIS_PASSWORD:-}
      - JWT_SECRET=${JWT_SECRET:-your_jwt_secret}
      - ADMIN_USERNAME=${ADMIN_USERNAME:-admin}
      - ADMIN_PASSWORD=${ADMIN_PASSWORD:-admin123}
      - BOT_TOKEN=${BOT_TOKEN}
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8000/"]
      interval: 30s
      timeout: 10s
      retries: 3

  postgres:
    image: postgres:15-alpine
    container_name: tgbot_postgres
    restart: unless-stopped
    environment:
      - POSTGRES_USER=${DB_USER:-tgbot}
      - POSTGRES_PASSWORD=${DB_PASSWORD:-tgbot123}
      - POSTGRES_DB=${DB_NAME:-tgbot}
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./migrations:/docker-entrypoint-initdb.d:ro
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${DB_USER:-tgbot} -d ${DB_NAME:-tgbot}"]
      interval: 10s
      timeout: 5s
      retries: 5

  redis:
    image: redis:7-alpine
    container_name: tgbot_redis
    restart: unless-stopped
    command: redis-server --appendonly yes ${REDIS_PASSWORD:+--requirepass $REDIS_PASSWORD}
    volumes:
      - redis_data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5

volumes:
  postgres_data:
  redis_data:
EOF
    echo -e "${GREEN}✓ Docker Hub 部署配置已创建${NC}"
}

# Deploy from Docker Hub
deploy_from_hub() {
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${CYAN}       从 Docker Hub 部署 (推荐)        ${NC}"
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""

    check_docker
    check_docker_compose
    check_env
    check_bot_token

    # Create hub compose file
    create_hub_compose

    echo -e "${YELLOW}拉取最新镜像...${NC}"
    docker pull $DOCKER_HUB_IMAGE:latest

    echo -e "${YELLOW}启动服务...${NC}"
    cd "$SCRIPT_DIR"
    docker_compose -f docker-compose.hub.yml up -d

    show_success
}

# Deploy from source
deploy_from_source() {
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${CYAN}         从源码构建部署                  ${NC}"
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""

    check_docker
    check_docker_compose
    check_env
    check_bot_token

    echo -e "${YELLOW}构建镜像...${NC}"
    cd "$SCRIPT_DIR"
    docker_compose build --no-cache

    echo -e "${YELLOW}启动服务...${NC}"
    docker_compose up -d

    show_success
}

# Quick start (auto-detect best method)
quick_start() {
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${CYAN}            快速部署                     ${NC}"
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""

    check_docker
    check_docker_compose
    check_env
    check_bot_token

    # Check if docker-compose.hub.yml exists
    if [ ! -f "$SCRIPT_DIR/docker-compose.hub.yml" ]; then
        create_hub_compose
    fi

    # Prefer Docker Hub if not building locally
    if [ -f "$SCRIPT_DIR/docker-compose.yml" ] && docker images | grep -q "tgbot-admin"; then
        echo -e "${YELLOW}使用本地镜像启动...${NC}"
        docker_compose up -d
    else
        echo -e "${YELLOW}从 Docker Hub 拉取镜像...${NC}"
        docker pull $DOCKER_HUB_API:latest
        docker pull $DOCKER_HUB_BOT:latest
        docker_compose -f docker-compose.hub.yml up -d
    fi

    show_success
}

# Show success message
show_success() {
    echo ""
    echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${GREEN}          🎉 部署成功！                    ${NC}"
    echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
    echo -e "Web 面板: ${BLUE}http://localhost:${API_PORT:-8000}${NC}"
    echo -e "默认账号: ${YELLOW}admin / admin123${NC}"
    echo ""
    echo -e "${CYAN}Bot 命令：${NC}"
    echo "  /start   - 开始使用"
    echo "  /help    - 查看帮助"
    echo "  /config  - 配置管理"
    echo "  /webui   - 获取面板链接"
    echo ""
    echo -e "${CYAN}管理命令：${NC}"
    echo "  ./start.sh logs     - 查看日志"
    echo "  ./start.sh stop     - 停止服务"
    echo "  ./start.sh restart  - 重启服务"
    echo "  ./start.sh status   - 查看状态"
    echo "  ./start.sh update   - 更新服务"
    echo ""
}

# Stop services
stop() {
    echo -e "${YELLOW}停止服务...${NC}"
    cd "$SCRIPT_DIR"
    if [ -f "docker-compose.hub.yml" ]; then
        docker_compose -f docker-compose.hub.yml down
    fi
    docker_compose down 2>/dev/null || true
    echo -e "${GREEN}✓ 服务已停止${NC}"
}

# Restart services
restart() {
    stop
    echo ""
    quick_start
}

# Show status
status() {
    echo -e "${YELLOW}服务状态：${NC}"
    echo ""
    cd "$SCRIPT_DIR"
    docker_compose ps 2>/dev/null || docker ps --filter "name=tgbot"
}

# Show logs
logs() {
    local service=$1
    cd "$SCRIPT_DIR"
    if [ -n "$service" ]; then
        docker_compose logs -f --tail=100 "$service"
    else
        docker_compose logs -f --tail=100
    fi
}

# Update
update() {
    echo -e "${YELLOW}更新服务...${NC}"

    # Pull latest code
    if [ -d ".git" ]; then
        echo -e "${YELLOW}拉取最新代码...${NC}"
        git pull
    fi

    # Pull latest images
    echo -e "${YELLOW}拉取最新镜像...${NC}"
    docker pull $DOCKER_HUB_IMAGE:latest

    # Restart
    stop
    quick_start
}

# Clean up
clean() {
    echo -e "${RED}警告: 这将删除所有数据！${NC}"
    read -p "确认删除？(y/N): " confirm
    if [[ "$confirm" =~ ^[Yy]$ ]]; then
        cd "$SCRIPT_DIR"
        docker_compose down -v 2>/dev/null || true
        docker_compose -f docker-compose.hub.yml down -v 2>/dev/null || true
        echo -e "${GREEN}✓ 清理完成${NC}"
    else
        echo "已取消"
    fi
}

# Show help
show_help() {
    echo ""
    echo -e "${CYAN}用法: $0 [命令]${NC}"
    echo ""
    echo "命令:"
    echo "  install     - 从零环境部署（自动安装 Docker）"
    echo "  hub         - 从 Docker Hub 拉取部署（推荐）"
    echo "  build       - 从源码构建部署"
    echo "  start       - 快速启动（默认）"
    echo "  stop        - 停止服务"
    echo "  restart     - 重启服务"
    echo "  status      - 查看状态"
    echo "  logs [svc]  - 查看日志"
    echo "  update      - 更新服务"
    echo "  clean       - 清理所有数据"
    echo ""
    echo "示例:"
    echo "  $0 install    # 全新安装（会自动安装 Docker）"
    echo "  $0 hub        # 从 Docker Hub 快速部署"
    echo "  $0 logs api   # 查看 API 日志"
    echo ""
}

# Main
case "$1" in
    install|"")
        print_banner
        check_root
        quick_start
        ;;
    hub)
        print_banner
        check_root
        deploy_from_hub
        ;;
    build)
        print_banner
        check_root
        deploy_from_source
        ;;
    start)
        print_banner
        check_root
        quick_start
        ;;
    stop)
        stop
        ;;
    restart)
        restart
        ;;
    status)
        status
        ;;
    logs)
        logs "$2"
        ;;
    update)
        update
        ;;
    clean)
        clean
        ;;
    help|--help|-h)
        show_help
        ;;
    *)
        echo -e "${RED}未知命令: $1${NC}"
        show_help
        exit 1
        ;;
esac
