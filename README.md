# Go Chat - WebSocket 聊天室应用

一个基于 Go 和 WebSocket 技术的实时聊天室应用。

## 功能特性

- ✅ 用户注册和登录（基于 Session + Cookie）
- ✅ 创建聊天室
- ✅ 邀请成员加入聊天室（需要房间创建者权限）
- ✅ 实时消息推送（WebSocket）
- ✅ 消息持久化（PostgreSQL）
- ✅ 查看历史消息
- ✅ 房间成员管理
- ✅ 响应式设计（Tailwind CSS）

## 技术栈

### 后端
- Go 1.24
- Gorilla WebSocket - WebSocket 支持
- Gorilla Mux - HTTP 路由
- Gorilla Sessions - Session 管理
- PostgreSQL - 数据库
- bcrypt - 密码加密

### 前端
- HTML5
- Tailwind CSS - UI 样式
- 原生 JavaScript - WebSocket 客户端

## 项目结构

```
go-chat/
├── cmd/
│   └── server/
│       └── main.go              # 主程序入口
├── internal/
│   ├── database/
│   │   └── database.go          # 数据库连接和初始化
│   ├── handlers/
│   │   ├── auth.go              # 认证处理器
│   │   ├── room.go              # 房间处理器
│   │   ├── member.go            # 成员管理处理器
│   │   └── websocket.go         # WebSocket 处理器
│   ├── middleware/
│   │   └── auth.go              # 认证中间件
│   ├── models/
│   │   └── models.go            # 数据模型
│   └── services/
│       └── hub/
│           ├── hub.go           # WebSocket Hub
│           └── connection.go    # WebSocket 连接处理
├── web/
│   ├── templates/
│   │   ├── login.html           # 登录页面
│   │   ├── register.html        # 注册页面
│   │   ├── rooms.html           # 房间列表页面
│   │   └── room.html            # 聊天室页面
│   └── static/
│       ├── css/
│       └── js/
├── migrations/
│   └── 001_init.sql             # 数据库迁移文件
├── go.mod
├── go.sum
├── .env.example                 # 环境变量示例
└── README.md
```

## 安装和运行

### 前置要求

1. Go 1.24+
2. PostgreSQL 12+

### 步骤

1. **克隆项目**（如果需要）

2. **安装依赖**
   ```bash
   go mod download
   ```

3. **配置数据库**

   创建 PostgreSQL 数据库：
   ```sql
   CREATE DATABASE gochat;
   ```

4. **配置环境变量**

   复制 `.env.example` 到 `.env` 并修改配置：
   ```bash
   cp .env.example .env
   ```

   修改 `.env` 文件中的数据库连接信息。

5. **运行应用**
   ```bash
   go run cmd/server/main.go
   ```

   应用将在 `http://localhost:8080` 启动。

## 使用说明

### 1. 注册账号

访问 `http://localhost:8080/register` 注册新账号。

### 2. 登录

使用注册的账号在 `http://localhost:8080/login` 登录。

### 3. 创建聊天室

登录后会进入房间列表页面，点击"创建新房间"按钮创建聊天室。

### 4. 邀请成员

进入聊天室后，点击"邀请成员"按钮，输入要邀请的用户名。

### 5. 聊天

在聊天室中输入消息并发送，消息会实时推送给房间内的所有成员。

## 数据库表结构

### users - 用户表
- id (主键)
- username (用户名，唯一)
- email (邮箱，唯一)
- password_hash (密码哈希)
- created_at, updated_at

### rooms - 聊天室表
- id (主键)
- name (房间名称)
- description (房间描述)
- creator_id (创建者 ID)
- created_at, updated_at

### room_members - 房间成员表
- id (主键)
- room_id (房间 ID)
- user_id (用户 ID)
- role (角色: creator/member)
- joined_at

### messages - 消息表
- id (主键)
- room_id (房间 ID)
- user_id (用户 ID)
- content (消息内容)
- created_at

### sessions - 会话表
- id (Session ID)
- user_id (用户 ID)
- data (会话数据)
- expires_at (过期时间)
- created_at

## API 端点

### 认证
- `POST /api/register` - 用户注册
- `POST /api/login` - 用户登录
- `GET /logout` - 退出登录

### 房间
- `GET /rooms` - 房间列表页面
- `GET /rooms/{id}` - 聊天室页面
- `POST /api/rooms` - 创建房间
- `GET /api/rooms/{id}/members` - 获取房间成员
- `POST /api/rooms/{id}/invite` - 邀请成员
- `DELETE /api/rooms/{id}/members/{memberId}` - 移除成员
- `POST /api/rooms/{id}/leave` - 离开房间

### WebSocket
- `GET /ws/rooms/{id}` - WebSocket 连接

## WebSocket 消息格式

### 客户端发送
```json
{
  "type": "message",
  "content": "消息内容"
}
```

### 服务端推送
```json
{
  "type": "message",
  "message": {
    "id": 1,
    "room_id": 1,
    "user_id": 1,
    "username": "张三",
    "content": "消息内容",
    "created_at": "2025-01-01T12:00:00Z"
  }
}
```

其他消息类型：
- `join` - 用户加入
- `leave` - 用户离开
- `error` - 错误消息

## 安全注意事项

⚠️ **生产环境部署前请注意：**

1. 修改 Session 密钥（在 `internal/handlers/auth.go` 和 `internal/middleware/auth.go` 中）
2. 启用 HTTPS
3. 配置 WebSocket 的 Origin 检查
4. 使用环境变量管理敏感配置
5. 添加速率限制
6. 添加 CSRF 保护

## 开发计划

- [ ] 添加文件上传功能
- [ ] 添加表情支持
- [ ] 添加消息撤回功能
- [ ] 添加在线状态显示
- [ ] 添加消息已读状态
- [ ] 添加私聊功能
- [ ] 优化性能和可扩展性

## 许可证

MIT License

## 贡献

欢迎提交 Issue 和 Pull Request！
