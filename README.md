# TGBot Admin - Telegram Bot 管理系统

分布式 Telegram Bot 管理系统，包含 Web 管理面板、高性能 Go API 后端、Python Bot 引擎。

## 功能特性

- **All-in-One 容器** - API + Bot + Web UI 合并到单个 Docker 镜像
- **多平台支持** - Linux/Windows/macOS (amd64/arm64)
- **自动发布** - 推送 Tag 自动创建 GitHub Release
- **Web 管理面板** - Tailwind CSS + Alpine.js 构建的响应式 SPA
- **高性能 API** - Go (Gin) 实现的 REST API + WebSocket
- **Bot 引擎** - Python + python-telegram-bot 异步处理
- **数据存储** - PostgreSQL + Redis 缓存
- **实时通信** - WebSocket 支持实时事件推送和指标监控

## 架构

```
┌─────────────────────────────────────────────────────────┐
│                 Web Dashboard (Tailwind)                │
└───────────────────────────┬─────────────────────────────┘
                            │ HTTP/WebSocket
┌───────────────────────────▼─────────────────────────────┐
│                   Go API Server (Gin)                   │
│              高并发 REST API + WebSocket                 │
└───────────────────────────┬─────────────────────────────┘
                            │
┌───────────────────────────┴─────────────────────────────┐
│  ┌──────────────────┐    ┌──────────────────┐           │
│  │   PostgreSQL     │    │      Redis       │           │
│  └──────────────────┘    └──────────────────┘           │
└───────────────────────────┬─────────────────────────────┘
                            │
┌───────────────────────────▼─────────────────────────────┐
│            Bot Engine (Python + python-telegram-bot)    │
└─────────────────────────────────────────────────────────┘
```

## 快速部署

### 方式一：从 Docker Hub 拉取（推荐）

```bash
# 1. 下载配置文件
curl -O https://raw.githubusercontent.com/nodesire7/TGBot_Admin/main/docker-compose.hub.yml
curl -O https://raw.githubusercontent.com/nodesire7/TGBot_Admin/main/.env.example
mv .env.example .env

# 2. 编辑 .env，填入 BOT_TOKEN
vim .env

# 3. 启动服务
docker-compose -f docker-compose.hub.yml up -d
```

### 方式二：从源码构建

```bash
# 1. 克隆项目
git clone https://github.com/nodesire7/TGBot_Admin.git
cd TGBot_Admin

# 2. 一键启动
./start.sh
```

### 方式三：下载二进制直接运行

从 [Releases](https://github.com/nodesire7/TGBot_Admin/releases) 页面下载对应平台的二进制文件。

```bash
# Linux/macOS
tar -xzf tgbot-admin-linux-amd64.tar.gz
./tgbot-admin

# Windows
# 解压 tgbot-admin-windows-amd64.zip
tgbot-admin.exe
```

## 启动脚本命令

```bash
./start.sh          # 启动服务（默认）
./start.sh stop     # 停止服务
./start.sh restart  # 重启服务
./start.sh status   # 查看状态
./start.sh logs     # 查看日志
./start.sh build    # 构建镜像
./start.sh clean    # 清理所有数据
./start.sh update   # 拉取更新并重启
```

## 多平台支持

| 平台 | 架构 | 文件 |
|------|------|------|
| Linux | amd64 | tgbot-admin-linux-amd64.tar.gz |
| Linux | arm64 | tgbot-admin-linux-arm64.tar.gz |
| Windows | amd64 | tgbot-admin-windows-amd64.zip |
| macOS | amd64 | tgbot-admin-darwin-amd64.tar.gz |
| macOS | arm64 (M1/M2) | tgbot-admin-darwin-arm64.tar.gz |

## Bot 指令

| 指令 | 说明 |
|------|------|
| `/start` | 开始使用机器人 |
| `/help` | 查看帮助信息 |
| `/config` | 打开配置面板（交互式按钮） |
| `/settimeout <秒>` | 设置验证超时时间 |
| `/setdifficulty <easy/medium/hard>` | 设置题目难度 |
| `/setmaxfail <次数>` | 设置最大失败次数 |
| `/kickonfail <on/off>` | 开启/关闭失败自动踢出 |
| `/autoapprove <on/off>` | 开启/关闭自动审批 |
| `/ban <用户ID> [原因]` | 封禁用户 |
| `/unban <用户ID>` | 解封用户 |
| `/blacklist` | 查看黑名单 |
| `/stats` | 查看统计数据 |
| `/webui` | 获取 Web 管理面板链接 |

## 项目结构

```
├── api/                    # Go API 服务
│   ├── main.go             # 入口
│   ├── config/             # 配置管理
│   ├── routers/            # 路由
│   │   └── routes/         # API 处理器
│   ├── models/             # 数据模型
│   └── middleware/         # 中间件
├── bot/                    # Python Bot 引擎
│   ├── main.py             # 入口
│   ├── database.py         # 数据库操作
│   ├── redis_client.py     # Redis 客户端
│   └── handlers/           # 事件处理器
├── web/                    # 前端
│   └── index.html          # SPA 入口
├── docker/                 # Docker 配置
│   ├── Dockerfile         # All-in-One 镜像
│   ├── supervisord.conf   # 进程管理
│   └── entrypoint.sh       # 入口脚本
├── migrations/             # 数据库迁移
├── docker-compose.yml      # Docker 编排（源码构建）
├── docker-compose.hub.yml  # Docker 编排（Docker Hub）
├── start.sh                # 一键启动脚本
└── .env.example            # 环境变量模板
```

## 本地开发

### 环境要求

- Go 1.24+
- Python 3.12+
- PostgreSQL 15+
- Redis 7+

### 开发步骤

```bash
# API 开发
cd api
go mod tidy
go run main.go

# Bot 开发
cd bot
pip install -r requirements.txt
python main.py
```

## API 文档

### 认证

```http
POST /api/auth/login
Content-Type: application/json

{"username": "admin", "password": "admin123"}
```

### 主要接口

```http
# 仪表盘
GET /api/dashboard/stats
GET /api/dashboard/timeline

# 群组
GET /api/groups
PUT /api/groups/{chat_id}
DELETE /api/groups/{chat_id}
POST /api/groups/{chat_id}/sync
GET /api/groups/{chat_id}/blacklist
POST /api/groups/{chat_id}/blacklist
DELETE /api/groups/{chat_id}/blacklist/{user_id}

# 插件
GET /api/plugins
PUT /api/plugins/{plugin_id}
POST /api/plugins/{plugin_id}/reload

# 日志
GET /api/logs/verification
GET /api/logs/action
```

### WebSocket

```javascript
// 事件流
ws://localhost:8000/ws/events

// 实时指标
ws://localhost:8000/ws/metrics
```

## Docker 镜像

本项目使用单容器部署：

```
docker pull nodesire7/tgbot-admin:latest
```

支持平台: `linux/amd64`, `linux/arm64`

容器内使用 supervisor 管理 API 和 Bot 进程。

### 多实例部署

不设置 `container_name`，Docker Compose 自动生成唯一名称，支持同一主机部署多个实例：

```bash
# 实例 1：在 /opt/tgbot1 目录
cd /opt/tgbot1
./start.sh

# 实例 2：在 /opt/tgbot2 目录（不同端口）
cd /opt/tgbot2
echo "API_PORT=8001" >> .env
./start.sh
```

容器名称自动生成为：`tgbot1_postgres_1`、`tgbot2_postgres_1` 等，互不冲突。

### 容器间通信

**重要：容器间通信使用 service name！**

```yaml
# ✅ 正确
environment:
  - DB_HOST=postgres      # service name
  - REDIS_HOST=redis      # service name

# ❌ 错误
environment:
  - DB_HOST=tgbot_postgres
  - REDIS_HOST=tgbot_redis
```

## 技术栈

| 组件 | 技术 |
|------|------|
| API | Go 1.24, Gin, pgx, go-redis |
| Bot | Python 3.12, python-telegram-bot, asyncpg |
| 前端 | Tailwind CSS, Alpine.js |
| 数据库 | PostgreSQL 15 |
| 缓存 | Redis 7 |
| 容器 | Docker, Docker Compose, Supervisor |
| CI/CD | GitHub Actions |

## 自动发布

推送 Tag 时自动触发：

1. 构建多平台二进制文件
2. 构建 Docker 镜像并推送到 Docker Hub
3. 创建 GitHub Release 并附带二进制文件

```bash
# 创建并推送新版本
git tag v1.0.1
git push origin v1.0.1
```

## 更新日志

### v1.1.1 (2026-03-14)
- 移除固定 container_name，支持同一主机部署多实例
- 修复 Docker Compose 容器间通信说明（service name vs container_name）
- 优化启动日志输出

### v1.0.0 (2026-03-14)
- 首个正式版本发布
- 单容器部署（API + Bot + Web UI）
- 多平台二进制支持
- GitHub Actions 自动发布

## License

MIT
