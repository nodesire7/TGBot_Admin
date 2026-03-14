#!/bin/bash
set -e

CONFIG_FILE="/app/data/config.json"

# 检查是否已配置
is_configured() {
    if [ -f "$CONFIG_FILE" ]; then
        # 检查配置文件中是否标记为已配置
        if grep -q '"is_configured":true' "$CONFIG_FILE" 2>/dev/null; then
            return 0
        fi
    fi
    return 1
}

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
            echo "警告: $name 未在规定时间内就绪"
            return 1
        fi
        echo "等待 $name... ($attempt/$max_attempts)"
        sleep 2
        ((attempt++))
    done
    echo "$name 已就绪"
    return 0
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
    local db_host=$1
    local db_port=$2
    local db_user=$3
    local db_password=$4
    local db_name=$5

    echo "初始化数据库..."

    # 运行迁移脚本
    for migration in /app/migrations/*.sql; do
        if [ -f "$migration" ]; then
            echo "执行迁移: $migration"
            PGPASSWORD="$db_password" psql \
                -h "$db_host" \
                -p "$db_port" \
                -U "$db_user" \
                -d "$db_name" \
                -f "$migration" 2>/dev/null || true
        fi
    done

    echo "数据库初始化完成"
}

# 创建数据目录
mkdir -p /app/data

# 主流程
main() {
    echo "================================"
    echo "  TGBot Admin 启动中..."
    echo "================================"

    if is_configured; then
        echo "系统已配置，启动正常模式..."
        echo ""

        # 从配置文件读取并设置环境变量
        if [ -f "$CONFIG_FILE" ]; then
            # 解析 JSON 并设置环境变量（简单方式）
            export DB_HOST=$(grep -o '"db_host":"[^"]*"' "$CONFIG_FILE" | cut -d'"' -f4)
            export DB_PORT=$(grep -o '"db_port":"[^"]*"' "$CONFIG_FILE" | cut -d'"' -f4)
            export DB_USER=$(grep -o '"db_user":"[^"]*"' "$CONFIG_FILE" | cut -d'"' -f4)
            export DB_PASSWORD=$(grep -o '"db_password":"[^"]*"' "$CONFIG_FILE" | cut -d'"' -f4)
            export DB_NAME=$(grep -o '"db_name":"[^"]*"' "$CONFIG_FILE" | cut -d'"' -f4)
            export REDIS_HOST=$(grep -o '"redis_host":"[^"]*"' "$CONFIG_FILE" | cut -d'"' -f4)
            export REDIS_PORT=$(grep -o '"redis_port":"[^"]*"' "$CONFIG_FILE" | cut -d'"' -f4)
            export REDIS_PASSWORD=$(grep -o '"redis_password":"[^"]*"' "$CONFIG_FILE" | cut -d'"' -f4)
        fi

        echo "数据库: ${DB_HOST}:${DB_PORT:-5432}"
        echo "Redis: ${REDIS_HOST}:${REDIS_PORT:-6379}"
        echo ""

        # 等待数据库
        if [ -n "${DB_HOST}" ]; then
            if ! wait_for_service "${DB_HOST}" "${DB_PORT:-5432}" "PostgreSQL"; then
                echo "数据库连接失败，请检查配置"
            fi
        fi

        # 等待 Redis
        if [ -n "${REDIS_HOST}" ]; then
            if ! wait_for_service "${REDIS_HOST}" "${REDIS_PORT:-6379}" "Redis"; then
                echo "Redis连接失败，请检查配置"
            fi
        fi

        # 初始化数据库
        init_database "${DB_HOST}" "${DB_PORT:-5432}" "${DB_USER}" "${DB_PASSWORD}" "${DB_NAME}"

        # 替换环境变量
        replace_env_vars
    else
        echo "系统未配置，启动配置向导模式..."
        echo "请访问 http://localhost:${API_PORT:-8000} 完成配置"
        echo ""
        # 只启动 API（不启动 Bot）
        export SETUP_MODE=1
    fi

    # 创建日志目录
    mkdir -p /var/log/supervisor

    # 启动 supervisor
    echo "启动服务..."
    exec /usr/bin/supervisord -c /etc/supervisor/supervisord.conf
}

main "$@"
