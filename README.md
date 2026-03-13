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

## 一键启动

### 快速开始

```bash
# 1. 克隆项目
git clone https://github.com/nodesire7/TGBot_Admin.git
cd TGBot_Admin

# 2. 配置 Bot Token
# 编辑 .env 文件，填入你的 BOT_TOKEN
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
./start.sh build    # 重新构建镜像
./start.sh clean    # 清理所有数据
./start.sh update   # 拉取更新并重启
```

### 访问地址

- Web 面板: http://localhost:8000
- 默认账号: admin / admin123

## Bot 命令

### 群组管理命令

| 命令 | 说明 | 示例 |
|------|------|------|
| `/start` | 开始使用 / 帮助 | `/start` |
| `/help` | 查看帮助 | `/help` |
| `/status` | 查看群组状态 | `/status` |
| `/enable` | 启用验证 | `/enable` |
| `/disable` | 禁用验证 | `/disable` |
| `/config` | 查看配置（带按钮） | `/config` |
| `/webui` | 获取管理面板链接 | `/webui` |
| `/stats` | 查看统计 | `/stats` |

### 配置命令

| 命令 | 说明 | 示例 |
|------|------|------|
| `/settimeout <秒>` | 设置验证超时 | `/settimeout 300` |
| `/setdifficulty <level>` | 设置难度 | `/setdifficulty medium` |
| `/setmaxfail <次数>` | 设置最大失败次数 | `/setmaxfail 3` |
| `/kickonfail <on\|off>` | 失败是否踢出 | `/kickonfail on` |
| `/autoapprove <on\|off>` | 自动通过申请 | `/autoapprove off` |
| `/setconfig <key> <value>` | 通用配置 | `/setconfig difficulty hard` |

### 黑名单命令

| 命令 | 说明 | 示例 |
|------|------|------|
| `/ban <user_id> [reason]` | 封禁用户 | `/ban 123456 垃圾广告` |
| `/unban <user_id>` | 解封用户 | `/unban 123456` |
| `/blacklist` | 查看黑名单 | `/blacklist` |

### 其他命令

| 命令 | 说明 | 示例 |
|------|------|------|
| `/resetstats confirm` | 清空日志（危险） | `/resetstats confirm` |

## 交互式配置

在群组中使用 `/config` 命令，会显示带按钮的配置面板：

```
⚙️ 群组配置管理

当前配置：
• 验证超时：300 秒
• 题目难度：easy
• 最大失败次数：3
• 失败踢出：是
• 自动通过：否

[验证超时: 300秒]
[难度: easy] [最大失败: 3次]
[失败踢出: ✅] [自动通过: ❌]
[🔄 重置为默认]
```

点击按钮即可快速切换配置值。

## Web 面板功能

### 概览页面
- Bot 在线状态、内存、CPU
- 实时事件流水线
- 今日验证统计

### 群组管理
- 群组列表搜索
- 一键启用/禁用
- 配置详情编辑
- 黑名单管理

### 插件管理
- 功能开关切换
- 热重载插件

### 日志查询
- 验证记录查询
- 操作日志查询

## 配置说明

### 验证配置项

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `verification_timeout` | int | 300 | 验证超时（秒），范围 30-3600 |
| `difficulty` | string | easy | 难度: easy/medium/hard |
| `max_fail_count` | int | 3 | 最大失败次数，范围 1-20 |
| `kick_on_fail` | bool | true | 验证失败是否踢出 |
| `auto_approve` | bool | false | 自动通过入群申请 |

### 难度说明

| 难度 | 题目类型 |
|------|----------|
| easy | 个位数加减法 |
| medium | 两位数加减法 |
| hard | 两位数乘除法 |

## 开发

### 目录结构

```
TGBot_Admin/
├── api/                    # Go API 服务
│   ├── main.go             # 入口
│   ├── config/             # 配置管理
│   ├── routers/            # 路由
│   ├── models/             # 数据模型
│   └── middleware/         # 中间件
│
├── bot/                    # Python Bot 引擎
│   ├── main.py             # 入口
│   ├── database.py         # 数据库操作
│   ├── redis_client.py     # Redis 客户端
│   └── handlers/           # 事件处理器
│
├── web/                    # 前端
│   └── index.html          # SPA 入口
│
├── migrations/             # 数据库迁移
├── docker-compose.yml      # Docker 编排
├── Dockerfile.api          # API 镜像
├── Dockerfile.bot          # Bot 镜像
├── start.sh                # 一键启动脚本
└── .env.example            # 环境变量模板
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
GET /api/groups/{chat_id}/blacklist

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

## License

MIT
