# TGBot_Admin - Telegram Bot 管理系统架构规划

## 1. 系统架构概览

```
┌─────────────────────────────────────────────────────────────────┐
│                        用户界面层                                │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │              Web Dashboard (Tailwind CSS)                │   │
│  │   Dashboard │ Groups │ Plugins │ Logs │ Settings        │   │
│  └─────────────────────────────────────────────────────────┘   │
└───────────────────────────┬─────────────────────────────────────┘
                            │ HTTP/WebSocket
┌───────────────────────────▼─────────────────────────────────────┐
│                     All-in-One 容器                             │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │              Go API Server (Gin)                        │   │
│  │   REST API │ WebSocket │ Auth │ Business Logic          │   │
│  └─────────────────────────────────────────────────────────┘   │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │              Bot Engine (python-telegram-bot)            │   │
│  │   验证模块 │ 禁言模块 │ 踢人模块 │ 消息处理               │   │
│  └─────────────────────────────────────────────────────────┘   │
│                      Supervisor 进程管理                         │
└───────────────────────────┬─────────────────────────────────────┘
                            │
┌───────────────────────────┴─────────────────────────────────────┐
│                        数据存储层                                │
│  ┌──────────────────┐    ┌──────────────────┐                  │
│  │   PostgreSQL     │    │      Redis       │                  │
│  │ (持久化配置/日志) │    │ (缓存/实时状态)   │                  │
│  └──────────────────┘    └──────────────────┘                  │
└─────────────────────────────────────────────────────────────────┘
```

---

## 2. 技术栈选型

| 层级 | 技术 | 版本要求 | 说明 |
|------|------|----------|------|
| 前端 | HTML + Tailwind CSS + Alpine.js | Tailwind 3.x | 简约现代 UI，轻量级响应式交互 |
| 后端 | Go (Gin) | 1.24+ | 高性能 REST API + WebSocket |
| 机器人 | python-telegram-bot | 20.x | 异步 Telegram Bot 框架 |
| 数据库 | PostgreSQL | 15+ | 主数据存储，支持 JSONB |
| 缓存 | Redis | 7+ | 会话管理、实时状态、消息队列 |
| 部署 | Docker + Docker Compose + Supervisor | - | 单容器化部署 |
| CI/CD | GitHub Actions | - | 自动构建、发布 |

---

## 3. 功能模块详细设计

### 3.1 概览工作台 (Dashboard Overview)

#### 3.1.1 运行状态卡片

| 指标 | 数据来源 | 更新频率 |
|------|----------|----------|
| Bot 在线状态 | Redis `bot:status` | 实时 (WebSocket) |
| 内存占用 | Bot 定时上报至 Redis | 5s |
| CPU 使用率 | Bot 定时上报至 Redis | 5s |
| 在线群组数 | PostgreSQL `groups` 表 | 10s |
| 累计拦截数 | PostgreSQL `logs` 表聚合 | 10s |
| 今日验证数 | Redis 计数器 `stats:today` | 实时 |

#### 3.1.2 实时流水线

```
数据流: Bot Engine → Redis Stream → FastAPI WebSocket → Frontend
存储: PostgreSQL logs 表 (持久化) + Redis Stream (实时推送)
```

**流水线事件类型**:
- `verification.success` - 验证成功
- `verification.failed` - 验证失败
- `verification.timeout` - 验证超时
- `user.kicked` - 用户被踢
- `user.muted` - 用户被禁言
- `user.banned` - 用户被封禁
- `group.joined` - Bot 加入新群组
- `group.left` - Bot 离开群组

### 3.2 群组管理中心 (Group Management)

#### 3.2.1 群组列表视图

**展示字段**:
- 群组头像 (Telegram Chat Photo)
- 群组名称
- 成员数量 (定时同步)
- Bot 状态 (启用/禁用)
- 今日验证统计 (成功/失败)
- 最后活跃时间

**操作按钮**:
- 进入详情
- 快速启用/禁用
- 同步信息

#### 3.2.2 群组详情页

**配置项**:

| 配置项 | 字段名 | 类型 | 默认值 | 说明 |
|--------|--------|------|--------|------|
| 启用状态 | `is_active` | boolean | true | 该群组是否启用 Bot |
| 验证超时 | `verification_timeout` | int | 300 | 验证时间限制 (秒) |
| 题目难度 | `difficulty` | enum | easy | easy(个位)/medium(十位)/hard(百位) |
| 自动同意申请 | `auto_approve` | boolean | false | 自动同意入群申请 |
| 失败踢出 | `kick_on_fail` | boolean | true | 验证失败是否踢出 |
| 失败次数限制 | `max_fail_count` | int | 3 | 最大失败次数 |
| 管理员白名单 | `admin_whitelist` | array | [] | 不需要验证的管理员列表 |

**黑名单管理**:
- 查看被封禁用户列表
- 添加/移除黑名单
- 封禁原因记录
- 封禁时间

### 3.3 插件/模块管理 (Plugin Management)

#### 3.3.1 插件列表

| 插件名称 | 插件ID | 状态 | 优先级 | 说明 |
|----------|--------|------|--------|------|
| 算术验证 | `arithmetic_verification` | 核心 | 1 | 必装核心模块 |
| 关键词屏蔽 | `keyword_filter` | 可选 | 2 | 消息关键词过滤 |
| 入群欢迎 | `welcome_message` | 可选 | 3 | 新成员欢迎消息 |
| 防洪水 | `flood_protection` | 可选 | 4 | 消息频率限制 |
| 链接过滤 | `link_filter` | 可选 | 5 | 链接检测过滤 |

#### 3.3.2 插件操作

- **启用/禁用**: 实时生效，写入数据库
- **配置编辑**: 弹窗表单，动态加载配置项
- **热重载**: 发送信号至 Bot 进程，重新加载插件模块
- **日志查看**: 跳转至该插件相关日志

---

## 4. 数据模型设计

### 4.1 PostgreSQL 表结构

#### 4.1.1 群组配置表 (groups)

```sql
CREATE TABLE groups (
    id SERIAL PRIMARY KEY,
    chat_id BIGINT UNIQUE NOT NULL,
    title VARCHAR(255) NOT NULL,
    username VARCHAR(255),
    description TEXT,
    member_count INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    config JSONB DEFAULT '{
        "verification_timeout": 300,
        "difficulty": "easy",
        "auto_approve": false,
        "kick_on_fail": true,
        "max_fail_count": 3
    }'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_active_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_groups_chat_id ON groups(chat_id);
CREATE INDEX idx_groups_is_active ON groups(is_active);
```

#### 4.1.2 用户黑名单表 (blacklist)

```sql
CREATE TABLE blacklist (
    id SERIAL PRIMARY KEY,
    chat_id BIGINT NOT NULL REFERENCES groups(chat_id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL,
    username VARCHAR(255),
    first_name VARCHAR(255),
    reason TEXT,
    banned_by BIGINT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(chat_id, user_id)
);

CREATE INDEX idx_blacklist_chat_id ON blacklist(chat_id);
CREATE INDEX idx_blacklist_user_id ON blacklist(user_id);
```

#### 4.1.3 验证记录表 (verification_logs)

```sql
CREATE TABLE verification_logs (
    id SERIAL PRIMARY KEY,
    chat_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    username VARCHAR(255),
    first_name VARCHAR(255),
    status VARCHAR(20) NOT NULL, -- success, failed, timeout
    question TEXT,
    answer TEXT,
    user_answer TEXT,
    attempt_count INTEGER DEFAULT 1,
    duration_seconds INTEGER,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_verification_logs_chat_id ON verification_logs(chat_id);
CREATE INDEX idx_verification_logs_created_at ON verification_logs(created_at);
CREATE INDEX idx_verification_logs_status ON verification_logs(status);
```

#### 4.1.4 插件配置表 (plugins)

```sql
CREATE TABLE plugins (
    id SERIAL PRIMARY KEY,
    plugin_id VARCHAR(100) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    is_enabled BOOLEAN DEFAULT TRUE,
    priority INTEGER DEFAULT 0,
    config JSONB DEFAULT '{}'::jsonb,
    last_restart_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

#### 4.1.5 操作日志表 (action_logs)

```sql
CREATE TABLE action_logs (
    id SERIAL PRIMARY KEY,
    chat_id BIGINT,
    user_id BIGINT,
    action_type VARCHAR(50) NOT NULL,
    action_data JSONB,
    operator_id BIGINT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_action_logs_created_at ON action_logs(created_at);
CREATE INDEX idx_action_logs_action_type ON action_logs(action_type);
```

#### 4.1.6 管理员表 (admins)

```sql
CREATE TABLE admins (
    id SERIAL PRIMARY KEY,
    user_id BIGINT UNIQUE NOT NULL,
    username VARCHAR(255),
    role VARCHAR(20) DEFAULT 'admin', -- super_admin, admin, viewer
    permissions JSONB DEFAULT '[]'::jsonb,
    is_active BOOLEAN DEFAULT TRUE,
    last_login_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

### 4.2 Redis 数据结构

| Key | 类型 | TTL | 说明 |
|-----|------|-----|------|
| `bot:status` | Hash | - | Bot 运行状态 (online, pid, started_at) |
| `bot:metrics` | Hash | 10s | 内存、CPU 等指标 |
| `verification:{chat_id}:{user_id}` | Hash | 5min | 当前验证会话数据 |
| `stats:today:{type}` | String | 24h | 今日统计计数器 |
| `stream:events` | Stream | 1h | 实时事件流 |
| `session:{token}` | Hash | 24h | WebUI 登录会话 |
| `cache:group:{chat_id}` | Hash | 10min | 群组配置缓存 |

---

## 5. API 接口设计

### 5.1 认证接口

```
POST   /api/auth/login          # 管理员登录
POST   /api/auth/logout         # 退出登录
GET    /api/auth/me             # 获取当前用户信息
POST   /api/auth/refresh        # 刷新 Token
```

### 5.2 仪表盘接口

```
GET    /api/dashboard/stats     # 获取统计数据
GET    /api/dashboard/metrics   # 获取实时指标 (WebSocket)
GET    /api/dashboard/timeline  # 获取最近事件流
```

### 5.3 群组管理接口

```
GET    /api/groups              # 群组列表 (分页、筛选)
GET    /api/groups/{chat_id}    # 群组详情
PUT    /api/groups/{chat_id}    # 更新群组配置
POST   /api/groups/{chat_id}/sync    # 同步群组信息
DELETE /api/groups/{chat_id}    # 移除群组记录

# 黑名单管理
GET    /api/groups/{chat_id}/blacklist
POST   /api/groups/{chat_id}/blacklist
DELETE /api/groups/{chat_id}/blacklist/{user_id}
```

### 5.4 插件管理接口

```
GET    /api/plugins             # 插件列表
PUT    /api/plugins/{plugin_id} # 更新插件状态/配置
POST   /api/plugins/{plugin_id}/reload  # 热重载插件
```

### 5.5 日志查询接口

```
GET    /api/logs/verification   # 验证日志查询
GET    /api/logs/action         # 操作日志查询
GET    /api/logs/export         # 日志导出
```

### 5.6 WebSocket 接口

```
WS     /ws/events               # 实时事件推送
WS     /ws/metrics              # 实时指标推送
```

---

## 6. 前端页面结构

```
/
├── index.html              # 入口 SPA
├── css/
│   └── styles.css          # Tailwind 输出
├── js/
│   ├── app.js              # Alpine.js 主应用
│   ├── api.js              # API 请求封装
│   ├── ws.js               # WebSocket 连接管理
│   └── utils.js            # 工具函数
└── pages/
    ├── dashboard.html      # 概览工作台
    ├── groups.html         # 群组列表
    ├── group-detail.html   # 群组详情
    ├── plugins.html        # 插件管理
    ├── logs.html           # 日志查询
    └── settings.html       # 系统设置
```

---

## 7. 目录结构设计

```
TGBot_Admin/
├── Dockerfile                  # All-in-One 镜像
├── docker-compose.yml          # Docker 编排（源码构建）
├── docker-compose.hub.yml      # Docker 编排（Docker Hub）
├── .env.example                # 环境变量模板
├── README.md                   # 项目说明
├── start.sh                    # 一键部署脚本
├── deploy.sh                   # 快速部署脚本
│
├── docker/                     # Docker 配置
│   ├── supervisord.conf        # 进程管理配置
│   └── entrypoint.sh           # 容器入口脚本
│
├── bot/                        # 机器人引擎
│   ├── main.py                 # Bot 入口
│   ├── config.py               # 配置管理
│   ├── database.py             # 数据库连接
│   ├── redis_client.py         # Redis 连接
│   ├── handlers/               # 事件处理器
│   │   ├── __init__.py
│   │   ├── verification.py     # 验证处理
│   │   ├── member.py           # 成员事件
│   │   └── admin.py            # 管理命令
│   ├── plugins/                # 插件模块
│   │   ├── __init__.py
│   │   ├── base.py             # 插件基类
│   │   ├── arithmetic.py       # 算术验证插件
│   │   └── keyword_filter.py   # 关键词过滤插件
│   ├── models/                 # 数据模型
│   │   ├── __init__.py
│   │   ├── group.py
│   │   ├── user.py
│   │   └── log.py
│   └── utils/                  # 工具函数
│       ├── __init__.py
│       └── helpers.py
│
├── api/                        # Go API 后端
│   ├── main.go                 # API 入口
│   ├── config/                 # 配置管理
│   │   └── config.go
│   ├── routers/                # 路由模块
│   │   ├── router.go
│   │   └── routes/             # API 处理器
│   │       ├── auth.go
│   │       ├── dashboard.go
│   │       ├── groups.go
│   │       ├── plugins.go
│   │       ├── logs.go
│   │       └── websocket.go
│   ├── models/                 # 数据模型
│   │   └── models.go
│   └── middleware/             # 中间件
│       └── auth.go
│
├── web/                        # 前端界面
│   └── index.html              # SPA 入口页面
│
├── migrations/                 # 数据库迁移
│   └── 001_init.sql
│
└── .github/                    # GitHub Actions
    └── workflows/
        └── docker-build.yml    # CI/CD 工作流
```

---

## 8. 交互流程详解

### 8.1 验证流程

```
1. 用户加入群组
   └─→ Bot 收到 ChatMember 事件
        └─→ 查询 Redis 缓存群组配置
             └─→ 配置不存在则查 PostgreSQL 并缓存
                  └─→ 检查 is_active
                       ├─→ True: 创建验证会话
                       │    └─→ 写入 Redis verification:{chat_id}:{user_id}
                       │    └─→ 发送验证消息
                       │    └─→ 写入 PostgreSQL verification_logs
                       │    └─→ 发布事件到 Redis Stream
                       └─→ False: 跳过验证

2. 用户回答验证
   └─→ Bot 收到消息回调
        └─→ 验证答案
             ├─→ 正确: 移除验证会话，记录成功
             └─→ 错误: 记录失败次数
                  └─→ 达到上限: 踢出用户

3. WebUI 实时更新
   └─→ WebSocket 订阅 Redis Stream
        └─→ 推送事件到前端
             └─→ 更新流水线和统计卡片
```

### 8.2 配置更新流程

```
1. WebUI 修改配置
   └─→ PUT /api/groups/{chat_id}
        └─→ 更新 PostgreSQL
             └─→ 删除 Redis 缓存
                  └─→ 返回成功

2. Bot 感知配置
   └─→ 下次操作时查询
        └─→ Redis 缓存失效
             └─→ 重新加载最新配置
```

---

## 9. 安全设计

### 9.1 认证机制
- 管理员通过 Telegram OAuth 登录
- Session 存储在 Redis，24h 过期
- API 请求需携带 Bearer Token

### 9.2 权限控制
- `super_admin`: 全部权限
- `admin`: 群组管理、配置修改
- `viewer`: 只读访问

### 9.3 数据安全
- 敏感配置通过环境变量注入
- 数据库连接使用 SSL
- Redis 访问需要密码认证

---

## 10. 部署方案

### 10.1 单容器部署 (推荐)

使用 All-in-One Docker 镜像，包含 API + Bot + Web UI：

```yaml
version: '3.8'

services:
  tgbot-admin:
    image: nodesire7/tgbot-admin:latest
    restart: unless-stopped
    ports:
      - "8000:8000"
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

  postgres:
    image: postgres:15-alpine
    restart: unless-stopped
    environment:
      - POSTGRES_USER=${DB_USER:-tgbot}
      - POSTGRES_PASSWORD=${DB_PASSWORD:-tgbot123}
      - POSTGRES_DB=${DB_NAME:-tgbot}
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./migrations:/docker-entrypoint-initdb.d:ro
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U tgbot"]
      interval: 10s
      timeout: 5s
      retries: 5

  redis:
    image: redis:7-alpine
    restart: unless-stopped
    command: redis-server --appendonly yes
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
```

### 10.2 快速部署命令

```bash
# 方式一：从 Docker Hub 部署（推荐）
curl -O https://raw.githubusercontent.com/nodesire7/TGBot_Admin/main/docker-compose.hub.yml
curl -O https://raw.githubusercontent.com/nodesire7/TGBot_Admin/main/.env.example
mv .env.example .env
# 编辑 .env 填入 BOT_TOKEN
docker-compose -f docker-compose.hub.yml up -d

# 方式二：从源码构建
git clone https://github.com/nodesire7/TGBot_Admin.git
cd TGBot_Admin
./start.sh

# 方式三：下载二进制直接运行
# 从 Releases 页面下载对应平台的二进制文件
tar -xzf tgbot-admin-linux-amd64.tar.gz
./tgbot-admin
```

### 10.3 多平台支持

| 平台 | 架构 | 文件 |
|------|------|------|
| Linux | amd64 | tgbot-admin-linux-amd64.tar.gz |
| Linux | arm64 | tgbot-admin-linux-arm64.tar.gz |
| Windows | amd64 | tgbot-admin-windows-amd64.zip |
| macOS | amd64 | tgbot-admin-darwin-amd64.tar.gz |
| macOS | arm64 (M1/M2) | tgbot-admin-darwin-arm64.tar.gz |

### 10.4 容器架构

```
┌─────────────────────────────────────┐
│        Docker Container             │
│  ┌───────────────────────────────┐  │
│  │        Supervisor             │  │
│  │  (进程管理)                    │  │
│  └───────────┬───────────────────┘  │
│              │                       │
│  ┌───────────┴───────────┐          │
│  │                       │          │
│  ▼                       ▼          │
│ ┌─────────────┐   ┌─────────────┐   │
│ │  Go API     │   │ Python Bot  │   │
│ │  (Gin)      │   │ (telegram)  │   │
│ │  :8000      │   │             │   │
│ └─────────────┘   └─────────────┘   │
│                                     │
│ ┌─────────────────────────────────┐ │
│ │        Web UI (Tailwind)        │ │
│ │        静态文件                  │ │
│ └─────────────────────────────────┘ │
└─────────────────────────────────────┘
```

---

## 11. 开发路线图

### Phase 1: 基础架构 (Week 1-2)
- [ ] 项目初始化，目录结构搭建
- [ ] PostgreSQL 表结构设计与迁移脚本
- [ ] Redis 连接与基础数据结构
- [ ] Bot 核心框架搭建
- [ ] FastAPI 基础框架搭建

### Phase 2: 核心功能 (Week 3-4)
- [ ] Bot 算术验证插件开发
- [ ] Bot 与数据库/Redis 交互
- [ ] API 认证模块开发
- [ ] API 群组管理接口
- [ ] 前端基础框架搭建

### Phase 3: WebUI 开发 (Week 5-6)
- [ ] Dashboard 概览页面
- [ ] 群组管理页面
- [ ] 插件管理页面
- [ ] WebSocket 实时通信
- [ ] 日志查询页面

### Phase 4: 优化与部署 (Week 7-8)
- [ ] 性能优化与缓存策略
- [ ] 错误处理与日志完善
- [ ] Docker 部署配置
- [ ] 文档编写
- [ ] 测试与上线

---

## 12. 监控指标

### 12.1 Bot 指标
- 在线状态与运行时长
- 内存/CPU 使用率
- 消息处理速率
- 验证成功率

### 12.2 系统指标
- API 响应时间
- 数据库连接池状态
- Redis 内存使用
- WebSocket 连接数

---

## 13. 扩展规划

### 未来功能
- 多语言支持
- 数据统计报表导出
- 自定义验证题目模板
- 多 Bot 实例管理
- 告警通知系统
- API Rate Limiting
