# TGBot Admin - Telegram Bot 管理系统

分布式 Telegram Bot 管理系统，包含 Web 管理面板、高性能 Go API 后端、Python Bot 引擎。

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

## 快速开始

### 1. 环境准备

```bash
# 克隆项目
cd TGBot_Admin

# 复制环境变量
cp .env.example .env

# 编辑 .env 填入配置
# - BOT_TOKEN: Telegram Bot Token
# - JWT_SECRET: JWT 密钥
# - REDIS_PASSWORD: Redis 密码
```

### 2. 启动服务

```bash
# 启动所有服务
docker-compose up -d

# 查看日志
docker-compose logs -f

# 初始化数据库（自动执行）
# 首次启动会自动运行 migrations/001_init.sql
```

### 3. 访问面板

- Web 面板: http://localhost:8000
- 默认账号: admin / admin123

## 开发

### 目录结构

```
TGBot_Admin/
├── api/                    # Go API 服务
│   ├── main.go             # 入口
│   ├── config/             # 配置管理
│   ├── routers/            # 路由
│   ├── models/             # 数据模型
│   ├── middleware/         # 中间件
│   └── go.mod              # Go 依赖
│
├── bot/                    # Python Bot 引擎
│   ├── main.py             # 入口
│   ├── config.py           # 配置
│   ├── database.py         # 数据库操作
│   ├── redis_client.py     # Redis 客户端
│   └── handlers/           # 事件处理器
│       ├── verification.py # 验证逻辑
│       ├── member.py       # 成员事件
│       └── admin.py        # 管理命令
│
├── web/                    # 前端
│   └── index.html          # SPA 入口
│
├── migrations/             # 数据库迁移
│   └── 001_init.sql
│
├── docker-compose.yml
├── Dockerfile.api
├── Dockerfile.bot
└── .env.example
```

### 本地开发

```bash
# API 开发
cd api
go mod tidy
go run main.go

# Bot 开发
cd bot
pip install -r requirements.txt
python main.py

# 前端开发
# 直接用浏览器打开 web/index.html
# 或用任意静态服务器
```

## 功能

### Bot 命令

| 命令 | 说明 | 权限 |
|------|------|------|
| `/start` | 开始使用 / 帮助 | 所有人 |
| `/status` | 查看群组状态 | 管理员 |
| `/enable` | 启用验证 | 管理员 |
| `/disable` | 禁用验证 | 管理员 |
| `/config` | 查看配置 | 管理员 |
| `/ban <user_id> [reason]` | 封禁用户 | 管理员 |
| `/unban <user_id>` | 解封用户 | 管理员 |

### Web 面板

- **概览**: Bot 状态、实时流水线、统计卡片
- **群组管理**: 列表、搜索、启用/禁用、配置详情
- **插件管理**: 功能开关、热重载
- **日志查询**: 验证记录、操作日志

## API 文档

### 认证

```http
POST /api/auth/login
Content-Type: application/json

{"username": "admin", "password": "admin123"}
```

### 仪表盘

```http
GET /api/dashboard/stats
Authorization: Bearer <token>

GET /api/dashboard/timeline
Authorization: Bearer <token>
```

### 群组

```http
GET /api/groups?page=1&limit=20
Authorization: Bearer <token>

GET /api/groups/{chat_id}
PUT /api/groups/{chat_id}
DELETE /api/groups/{chat_id}

GET /api/groups/{chat_id}/blacklist
POST /api/groups/{chat_id}/blacklist
DELETE /api/groups/{chat_id}/blacklist/{user_id}
```

### 插件

```http
GET /api/plugins
PUT /api/plugins/{plugin_id}
POST /api/plugins/{plugin_id}/reload
```

### WebSocket

```javascript
// 事件流
ws://localhost:8000/ws/events

// 实时指标
ws://localhost:8000/ws/metrics
```

## 配置说明

### 群组配置

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `verification_timeout` | int | 300 | 验证超时（秒） |
| `difficulty` | string | easy | 难度: easy/medium/hard |
| `auto_approve` | bool | false | 自动同意入群申请 |
| `kick_on_fail` | bool | true | 验证失败踢出 |
| `max_fail_count` | int | 3 | 最大失败次数 |
| `admin_whitelist` | array | [] | 跳过验证的管理员 |

## 扩展插件

### 创建新插件

1. 在 `bot/plugins/` 创建插件文件:

```python
# bot/plugins/my_plugin.py
from plugins.base import BasePlugin

class MyPlugin(BasePlugin):
    plugin_id = "my_plugin"
    name = "我的插件"
    priority = 10

    async def on_message(self, update, context):
        # 处理消息
        pass
```

2. 在数据库添加插件记录:

```sql
INSERT INTO plugins (plugin_id, name, is_enabled, priority)
VALUES ('my_plugin', '我的插件', true, 10);
```

## License

MIT
