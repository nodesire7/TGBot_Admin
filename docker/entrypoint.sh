#!/bin/bash
set -e

# 等待依赖服务
wait_for_service() {
    local host=$1
    local port=$2
    local name=$3
    local max_attempts=30
    local attempt=1

    echo "等待 $name 就绪..."
    while ! nc -z "$host" "$port" 2>/dev/null; do
        if [ $attempt -ge $max_attempts ]; then
            echo "错误: $name 未在规定时间内就绪"
            exit 1
        fi
        echo "等待 $name... ($attempt/$max_attempts)"
        sleep 2
        ((attempt++))
    done
    echo "$name 已就绪"
}

# 替换 supervisor 配置中的环境变量
replace_env_vars() {
    local conf_file="/etc/supervisor/conf.d/tgbot-admin.conf"

    # 替换所有环境变量
    sed -i "s|%(ENV_DB_HOST)s|${DB_HOST:-postgres}|g" "$conf_file"
    sed -i "s|%(ENV_DB_PORT)s|${DB_PORT:-5432}|g" "$conf_file"
    sed -i "s|%(ENV_DB_USER)s|${DB_USER:-tgbot}|g" "$conf_file"
    sed -i "s|%(ENV_DB_PASSWORD)s|${DB_PASSWORD:-tgbot123}|g" "$conf_file"
    sed -i "s|%(ENV_DB_NAME)s|${DB_NAME:-tgbot}|g" "$conf_file"
    sed -i "s|%(ENV_REDIS_HOST)s|${REDIS_HOST:-redis}|g" "$conf_file"
    sed -i "s|%(ENV_REDIS_PORT)s|${REDIS_PORT:-6379}|g" "$conf_file"
    sed -i "s|%(ENV_REDIS_PASSWORD)s|${REDIS_PASSWORD:-}|g" "$conf_file"
    sed -i "s|%(ENV_JWT_SECRET)s|${JWT_SECRET:-change_me}|g" "$conf_file"
    sed -i "s|%(ENV_ADMIN_USERNAME)s|${ADMIN_USERNAME:-admin}|g" "$conf_file"
    sed -i "s|%(ENV_ADMIN_PASSWORD)s|${ADMIN_PASSWORD:-admin123}|g" "$conf_file"
    sed -i "s|%(ENV_BOT_TOKEN)s|${BOT_TOKEN:-}|g" "$conf_file"
}

# 初始化数据库
init_database() {
    echo "初始化数据库..."

    # 运行迁移脚本
    for migration in /app/migrations/*.sql; do
        if [ -f "$migration" ]; then
            echo "执行迁移: $migration"
            PGPASSWORD="${DB_PASSWORD:-tgbot123}" psql \
                -h "${DB_HOST:-postgres}" \
                -p "${DB_PORT:-5432}" \
                -U "${DB_USER:-tgbot}" \
                -d "${DB_NAME:-tgbot}" \
                -f "$migration" 2>/dev/null || true
        fi
    done

    echo "数据库初始化完成"
}

# 主流程
main() {
    echo "================================"
    echo "  TGBot Admin 启动中..."
    echo "================================"

    # 显示连接信息
    echo "数据库: ${DB_HOST:-postgres}:${DB_PORT:-5432}"
    echo "Redis: ${REDIS_HOST:-redis}:${REDIS_PORT:-6379}"
    echo ""

    # 等待数据库
    if [ -n "${DB_HOST}" ]; then
        wait_for_service "${DB_HOST}" "${DB_PORT:-5432}" "PostgreSQL"
    fi

    # 等待 Redis
    if [ -n "${REDIS_HOST}" ]; then
        wait_for_service "${REDIS_HOST}" "${REDIS_PORT:-6379}" "Redis"
    fi

    # 初始化数据库
    init_database

    # 替换环境变量
    replace_env_vars

    # 创建日志目录
    mkdir -p /var/log/supervisor

    # 启动 supervisor
    echo "启动服务..."
    exec /usr/bin/supervisord -c /etc/supervisor/supervisord.conf
}

main "$@"
