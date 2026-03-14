# TGBot Admin - Telegram Bot 插件化管理平台

分布式 Telegram Bot 插件化管理平台，支持多 Bot 管理、插件系统、在线开发，包含 Web 管理面板、高性能 Go API 后端、Python Bot 引擎。

## 功能特性

- **多 Bot 管理** - 一套系统管理多个 Telegram Bot
- **插件系统** - 即插即用的插件架构，支持官方/社区/本地插件
- **在线开发** - 内置 Web IDE，支持在线开发和测试插件
- **插件市场** - 浏览和安装社区插件
- **Web 配置向导** - 首次启动通过 Web UI 完成配置，无需手动编辑文件
- **All-in-One 容器** - API + Bot + Web UI 合并到单个 Docker 镜像
- **多平台支持** - Linux/Windows/macOS (amd64/arm64)
- **自动发布** - 推送 Tag 自动创建 GitHub Release
- **Web 管理面板** - Tailwind CSS + Alpine.js 构建的响应式 SPA
- **高性能 API** - Go (Gin) 实现的 REST API + WebSocket
- **Bot 引擎** - Python + python-telegram-bot 异步处理
- **数据存储** - PostgreSQL + Redis 缓存
- **实时通信** - WebSocket 支持实时事件推送和指标监控

## 插件系统

系统内置 8 个官方插件：

| 插件 | 说明 | 默认启用 |
|------|------|---------|
| arithmetic_verification | 算术验证，新用户入群验证 | ✅ |
| welcome_message | 入群欢迎消息 | ✅ |
| keyword_filter | 关键词过滤 | ✅ |
| flood_protection | 防洪水攻击 | ✅ |
| link_filter | 链接过滤 | ✅ |
| anti_spam | 反垃圾消息 | ✅ |
| auto_reply | 自动回复 | ❌ |
| stats_reporter | 统计报告 | ❌ |

### 开发自定义插件

使用 Python SDK 快速开发插件：

```python
from tgbot_plugin import Plugin, Context, User, Message

class MyPlugin(Plugin):
    id = "my_plugin"
    name = "我的插件"
    version = "1.0.0"
    author = "Developer"

    @Plugin.on_join
    async def on_user_join(self, ctx: Context, user: User):
        await ctx.send_message(f"欢迎 {user.full_name}!")

    @Plugin.on_command("hello")
    async def hello_cmd(self, ctx: Context, args):
        await ctx.reply("Hello World!")

    @Plugin.on_message
    async def on_message(self, ctx: Context, msg: Message):
        if "敏感词" in msg.text:
            await msg.delete()
            await ctx.kick_user(msg.from_user.id)
```

### 插件钩子

| 钩子 | 触发时机 | 可用操作 |
|------|---------|---------|
| `on_join` | 用户入群 | 验证、欢迎、踢出 |
| `on_leave` | 用户退群 | 记录、通知 |
| `on_message` | 收到消息 | 过滤、回复、删除 |
| `on_command` | 命令触发 | 自定义命令处理 |
| `on_callback` | 按钮回调 | 处理交互 |
| `on_error` | 错误发生 | 日志、通知 |

### 插件权限

```python
from tgbot_plugin import Permission

class MyPlugin(Plugin):
    permissions = [
        Permission.SEND_MESSAGES,
        Permission.DELETE_MESSAGES,
        Permission.KICK_MEMBERS,
    ]
```

## 多 Bot 管理

系统支持在一个管理面板中管理多个 Telegram Bot：

- **添加 Bot** - 通过 Web UI 添加新的 Bot，自动验证 Token 有效性
- **Bot 状态** - 实时监控每个 Bot 的运行状态、CPU、内存占用
- **群组关联** - 每个 Bot 可以关联多个群组
- **独立配置** - 每个 Bot 可以有独立的验证配置（超时时间、难度等）
- **主 Bot 标记** - 设置主 Bot，用于系统级操作

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

# 2. 启动服务
docker-compose -f docker-compose.hub.yml up -d

# 3. 访问 Web 界面完成配置向导
# 打开 http://localhost:8000 按提示完成配置
```

**首次启动会进入配置向导，无需手动编辑配置文件：**
- 配置数据库连接（带测试按钮）
- 配置 Redis 连接（带测试按钮）
- 填写 Bot Token
- 设置管理员账号

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

# Bot 管理
GET /api/bots                  # 获取所有 Bot 列表
POST /api/bots                 # 添加新 Bot
GET /api/bots/{id}             # 获取 Bot 详情
PUT /api/bots/{id}             # 更新 Bot 配置
DELETE /api/bots/{id}          # 删除 Bot
POST /api/bots/{id}/start      # 启动 Bot
POST /api/bots/{id}/stop       # 停止 Bot
POST /api/bots/{id}/restart    # 重启 Bot
POST /api/bots/test-token      # 测试 Bot Token 有效性

# 群组
GET /api/groups
PUT /api/groups/{chat_id}
DELETE /api/groups/{chat_id}
POST /api/groups/{chat_id}/sync
GET /api/groups/{chat_id}/blacklist
POST /api/groups/{chat_id}/blacklist
DELETE /api/groups/{chat_id}/blacklist/{user_id}

# 插件管理
GET    /api/plugins                    # 已安装插件列表
POST   /api/plugins/install            # 安装插件
GET    /api/plugins/{id}               # 插件详情
PUT    /api/plugins/{id}               # 更新插件配置
DELETE /api/plugins/{id}               # 卸载插件
POST   /api/plugins/{id}/enable        # 启用插件
POST   /api/plugins/{id}/disable       # 禁用插件
POST   /api/plugins/{id}/test          # 测试插件
GET    /api/plugins/{id}/logs          # 插件执行日志

# Bot 插件配置
GET    /api/bots/{id}/plugins          # Bot 的插件配置
PUT    /api/bots/{id}/plugins/{pid}    # 更新 Bot 插件配置

# 在线开发
GET    /api/ide/templates              # 插件模板列表
POST   /api/ide/compile                # 编译检查
POST   /api/ide/run                    # 沙箱运行
POST   /api/ide/test                   # 模拟测试
POST   /api/ide/deploy                 # 部署插件
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

### v2.0.0 (2026-03-14)
- **插件系统重构**
  - 新增完整的插件生命周期管理 (install/load/enable/start/stop/disable/uninstall)
  - 新增 Hook 系统：on_join, on_message, on_command, on_callback 等
  - 新增 Python 插件 SDK (`bot/plugin_sdk/`)
  - 支持每个 Bot 独立配置插件
  - 插件沙箱执行环境 (规划中)
- **数据库扩展**
  - 新增 `plugins` 表存储插件
  - 新增 `bot_plugins` 表存储 Bot 插件配置
  - 新增 `plugin_logs` 表记录执行日志
  - 新增 `hook_registry` 表管理钩子注册
  - 新增 `user_plugins` 表支持在线开发
  - 新增 `market_plugins` 表缓存市场插件
- **内置官方插件**
  - arithmetic_verification: 算术验证
  - welcome_message: 入群欢迎
  - keyword_filter: 关键词过滤
  - flood_protection: 防洪水攻击
  - link_filter: 链接过滤
  - anti_spam: 反垃圾消息
  - auto_reply: 自动回复
  - stats_reporter: 统计报告

### v1.3.0 (2026-03-14)
- 新增多 Bot 管理：一套系统管理多个 Telegram Bot
- 添加 Bot 管理页面：添加、删除、启停 Bot
- Bot Token 验证：添加 Bot 时自动验证 Token 有效性
- 数据库新增 `bots` 表支持多 Bot 数据存储
- 群组支持关联到指定 Bot

### v1.2.0 (2026-03-14)
- 新增 Web 配置向导（类似 WordPress 安装流程）
- 支持通过 Web UI 配置数据库、Redis、Bot Token
- 配置页面带连接测试按钮
- 首次启动无需手动编辑配置文件
- 移除 container_name，支持多实例部署

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
