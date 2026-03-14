# TGBot Admin - Telegram Bot 管理系统

分布式 Telegram Bot 管理系统，包含 Web 管理面板、高性能 Go API 后端、Python Bot 引擎。

## 功能特性

- **Web 管理面板** - Tailwind CSS + Alpine.js 构建的响应式 SPA
- **高性能 API** - Go (Gin) 实现的 REST API + WebSocket
- **Bot 引擎** - Python + python-telegram-bot 异步处理
- **数据存储** - PostgreSQL + Redis 缓存
- **实时通信** - WebSocket 支持实时事件推送和指标监控
- **Docker 部署** - 支持多平台 (amd64/arm64) 容器化部署
- **CI/CD** - GitHub Actions 自动构建推送 Docker 镜像

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

# 首次运行会自动创建 .env 文件
# 请编辑 .env 填入从 @BotFather 获取的 Token
```

### 启动脚本命令

```bash
./start.sh          # 启动服务
./start.sh stop     # 停止服务
./start.sh restart  # 重启服务
./start.sh status   # 查看状态
./start.sh logs     # 查看日志
./start.sh build    # 构建镜像
./start.sh clean    # 清理容器和镜像
./start.sh update   # 拉取最新代码并重启
```

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
├── migrations/             # 数据库迁移
├── docker-compose.yml      # Docker 编排
├── Dockerfile.api          # API 镜像
├── Dockerfile.bot          # Bot 镜像
├── start.sh                # 一键启动脚本
└── .env.example            # 环境变量模板
```

## 本地开发

### 环境要求

- Go 1.24+
- Python 3.11+
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

本项目已发布到 Docker Hub：

- **API 镜像**: `docker.io/nodesire7/tgbot-admin-api:latest`
- **Bot 镜像**: `docker.io/nodesire7/tgbot-admin-bot:latest`

支持平台: `linux/amd64`, `linux/arm64`

## 技术栈

| 组件 | 技术 |
|------|------|
| API | Go 1.24, Gin, pgx, go-redis |
| Bot | Python 3.11, python-telegram-bot, asyncpg |
| 前端 | Tailwind CSS, Alpine.js |
| 数据库 | PostgreSQL 15 |
| 缓存 | Redis 7 |
| 容器 | Docker, Docker Compose |
| CI/CD | GitHub Actions |

## 更新日志

### 2026-03-14
- 修复 Go 代码中 `redis` 变量名与包名冲突问题
- 修复 Go 代码中未使用的 `context` 导入
- 更新 Dockerfile.api 使用 Go 1.24 并支持自动工具链升级
- 修复 Dockerfile.bot 中 psutil 编译依赖
- 优化 GitHub Actions 工作流配置

## License

MIT
