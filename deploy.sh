#!/bin/bash
# 快速部署脚本 - 从 Docker Hub 拉取并运行

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

print_banner() {
    echo -e "${BLUE}"
    echo "╔═══════════════════════════════════════════╗"
    echo "║      TGBot Admin - 快速部署脚本           ║"
    echo "╚═══════════════════════════════════════════╝"
    echo -e "${NC}"
}

check_dependencies() {
    if ! command -v docker &> /dev/null; then
        echo -e "${RED}错误: Docker 未安装${NC}"
        exit 1
    fi
}

check_env() {
    if [ ! -f ".env" ]; then
        cp .env.example .env
        JWT_SECRET=$(openssl rand -hex 32 2>/dev/null || echo "jwt_$(date +%s)")
        DB_PASSWORD=$(openssl rand -hex 16 2>/dev/null || echo "db_$(date +%s)")
        REDIS_PASSWORD=$(openssl rand -hex 16 2>/dev/null || echo "redis_$(date +%s)")

        if [[ "$OSTYPE" == "darwin"* ]]; then
            sed -i '' "s/your_super_secret_jwt_key_change_this_in_production/${JWT_SECRET}/g" .env
            sed -i '' "s/your_db_password_here/${DB_PASSWORD}/g" .env
            sed -i '' "s/your_redis_password/${REDIS_PASSWORD}/g" .env
        else
            sed -i "s/your_super_secret_jwt_key_change_this_in_production/${JWT_SECRET}/g" .env
            sed -i "s/your_db_password_here/${DB_PASSWORD}/g" .env
            sed -i "s/your_redis_password/${REDIS_PASSWORD}/g" .env
        fi

        echo -e "${YELLOW}✓ .env 已创建，请填入 BOT_TOKEN${NC}"
    fi
}

case "$1" in
    start|"")
        print_banner
        check_dependencies
        check_env
        echo -e "${YELLOW}启动服务...${NC}"
        docker-compose up -d
        echo -e "${GREEN}✓ 服务已启动${NC}"
        echo -e "Web 面板: ${BLUE}http://localhost:8000${NC}"
        ;;

    stop)
        docker-compose down
        echo -e "${GREEN}✓ 服务已停止${NC}"
        ;;

    restart)
        docker-compose down
        docker-compose up -d
        echo -e "${GREEN}✓ 服务已重启${NC}"
        ;;

    pull)
        echo -e "${YELLOW}拉取最新镜像...${NC}"
        docker-compose pull
        echo -e "${GREEN}✓ 镜像已更新${NC}"
        ;;

    logs)
        docker-compose logs -f --tail=100 ${2:-}
        ;;

    status)
        docker-compose ps
        ;;

    clean)
        echo -e "${RED}警告: 将删除所有数据${NC}"
        read -p "确认？(y/N): " confirm
        [ "$confirm" = "y" ] && docker-compose down -v && echo -e "${GREEN}✓ 已清理${NC}"
        ;;

    *)
        echo "用法: $0 {start|stop|restart|pull|logs|status|clean}"
        ;;
esac
