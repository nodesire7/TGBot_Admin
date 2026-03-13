#!/bin/bash

# TGBot Admin 一键启动脚本
# 支持: start, stop, restart, status, logs, build, clean

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Project directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# Print banner
print_banner() {
    echo -e "${BLUE}"
    echo "╔═══════════════════════════════════════════╗"
    echo "║         TGBot Admin - 管理系统            ║"
    echo "║     Go API + Python Bot + Tailwind UI     ║"
    echo "╚═══════════════════════════════════════════╝"
    echo -e "${NC}"
}

# Check dependencies
check_dependencies() {
    echo -e "${YELLOW}检查依赖...${NC}"

    if ! command -v docker &> /dev/null; then
        echo -e "${RED}错误: Docker 未安装${NC}"
        echo "请先安装 Docker: https://docs.docker.com/get-docker/"
        exit 1
    fi

    if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
        echo -e "${RED}错误: Docker Compose 未安装${NC}"
        echo "请先安装 Docker Compose"
        exit 1
    fi

    echo -e "${GREEN}✓ Docker 已安装${NC}"
    echo -e "${GREEN}✓ Docker Compose 已安装${NC}"
}

# Check if .env exists
check_env() {
    if [ ! -f ".env" ]; then
        echo -e "${YELLOW}.env 文件不存在，正在创建...${NC}"
        cp .env.example .env

        # Generate random secrets
        JWT_SECRET=$(openssl rand -hex 32 2>/dev/null || echo "jwt_secret_$(date +%s)")
        REDIS_PASSWORD=$(openssl rand -hex 16 2>/dev/null || echo "redis_pass_$(date +%s)")

        # Update .env with generated secrets
        if [[ "$OSTYPE" == "darwin"* ]]; then
            sed -i '' "s/your_super_secret_jwt_key_change_this_in_production/${JWT_SECRET}/g" .env
            sed -i '' "s/your_redis_password/${REDIS_PASSWORD}/g" .env
        else
            sed -i "s/your_super_secret_jwt_key_change_this_in_production/${JWT_SECRET}/g" .env
            sed -i "s/your_redis_password/${REDIS_PASSWORD}/g" .env
        fi

        echo -e "${GREEN}✓ .env 已创建${NC}"
        echo -e "${YELLOW}请编辑 .env 文件，填入您的 BOT_TOKEN${NC}"
        echo ""
    fi
}

# Check if BOT_TOKEN is set
check_bot_token() {
    if grep -q "your_bot_token_here" .env 2>/dev/null; then
        echo -e "${RED}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
        echo -e "${RED}错误: BOT_TOKEN 未配置！${NC}"
        echo -e "${RED}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
        echo ""
        echo "请按照以下步骤获取 Bot Token："
        echo "1. 在 Telegram 中搜索 @BotFather"
        echo "2. 发送 /newbot 命令创建新 Bot"
        echo "3. 按提示设置 Bot 名称"
        echo "4. 复制获得的 Token"
        echo ""
        echo "然后编辑 .env 文件："
        echo "  BOT_TOKEN=你的token"
        echo ""
        read -p "按 Enter 继续（已配置好 Token）或 Ctrl+C 退出: "
    fi
}

# Build containers
build() {
    echo -e "${YELLOW}构建 Docker 镜像...${NC}"
    docker-compose build --no-cache
    echo -e "${GREEN}✓ 镜像构建完成${NC}"
}

# Start services
start() {
    print_banner
    check_dependencies
    check_env
    check_bot_token

    echo -e "${YELLOW}启动服务...${NC}"
    docker-compose up -d

    echo ""
    echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${GREEN}✓ 服务启动成功！${NC}"
    echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
    echo -e "Web 面板: ${BLUE}http://localhost:8000${NC}"
    echo -e "默认账号: ${YELLOW}admin / admin123${NC}"
    echo ""
    echo "Bot 命令："
    echo "  /help     - 查看帮助"
    echo "  /status   - 群组状态"
    echo "  /config   - 配置管理"
    echo "  /webui    - 获取面板链接"
    echo ""
    echo "管理命令："
    echo "  ./start.sh logs     - 查看日志"
    echo "  ./start.sh stop     - 停止服务"
    echo "  ./start.sh restart  - 重启服务"
    echo "  ./start.sh status   - 查看状态"
}

# Stop services
stop() {
    echo -e "${YELLOW}停止服务...${NC}"
    docker-compose down
    echo -e "${GREEN}✓ 服务已停止${NC}"
}

# Restart services
restart() {
    stop
    echo ""
    start
}

# Show status
status() {
    echo -e "${YELLOW}服务状态：${NC}"
    echo ""
    docker-compose ps
}

# Show logs
logs() {
    local service=$1
    if [ -z "$service" ]; then
        docker-compose logs -f --tail=100
    else
        docker-compose logs -f --tail=100 "$service"
    fi
}

# Clean up
clean() {
    echo -e "${RED}警告: 这将删除所有数据！${NC}"
    read -p "确认删除？(y/N): " confirm
    if [ "$confirm" = "y" ] || [ "$confirm" = "Y" ]; then
        docker-compose down -v
        echo -e "${GREEN}✓ 清理完成${NC}"
    else
        echo "已取消"
    fi
}

# Update
update() {
    echo -e "${YELLOW}拉取最新代码...${NC}"
    git pull
    echo -e "${YELLOW}重新构建镜像...${NC}"
    docker-compose build --no-cache
    echo -e "${YELLOW}重启服务...${NC}"
    docker-compose up -d
    echo -e "${GREEN}✓ 更新完成${NC}"
}

# Main
case "$1" in
    start|"")
        start
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
    build)
        build
        ;;
    clean)
        clean
        ;;
    update)
        update
        ;;
    *)
        echo "用法: $0 {start|stop|restart|status|logs|build|clean|update}"
        echo ""
        echo "命令说明："
        echo "  start   - 启动服务（默认）"
        echo "  stop    - 停止服务"
        echo "  restart - 重启服务"
        echo "  status  - 查看状态"
        echo "  logs    - 查看日志 [service]"
        echo "  build   - 构建镜像"
        echo "  clean   - 清理所有数据"
        echo "  update  - 更新并重启"
        ;;
esac
