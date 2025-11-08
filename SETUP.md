# 安装和配置指南

## 第一步：安装 PostgreSQL

### Windows
1. 下载 PostgreSQL：https://www.postgresql.org/download/windows/
2. 运行安装程序，记住你设置的密码
3. 默认端口为 5432，默认用户为 postgres

### 验证安装
打开命令行，运行：
```bash
psql --version
```

## 第二步：创建数据库

### 方法 1：使用 pgAdmin（图形界面）
1. 打开 pgAdmin
2. 右键点击 "Databases" -> "Create" -> "Database"
3. 输入数据库名：`gochat`
4. 点击 "Save"

### 方法 2：使用命令行
```bash
# 登录 PostgreSQL
psql -U postgres

# 创建数据库
CREATE DATABASE gochat;

# 退出
\q
```

## 第三步：配置环境变量（可选）

如果你的数据库配置与默认值不同，请创建 `.env` 文件：

```bash
cp .env.example .env
```

然后编辑 `.env` 文件：
```
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=你的密码
DB_NAME=gochat
```

或者，你可以通过环境变量设置：
```bash
# Windows PowerShell
$env:DB_PASSWORD="你的密码"

# Windows CMD
set DB_PASSWORD=你的密码
```

## 第四步：运行应用

```bash
cd cmd/server
go run main.go
```

## 第五步：访问应用

打开浏览器访问：http://localhost:8080

## 常见问题

### 1. 数据库连接失败
**错误**: `failed to ping database: pq: password authentication failed`

**解决方案**:
- 检查 PostgreSQL 是否正在运行
- 确认数据库密码正确
- 设置环境变量 `DB_PASSWORD`

### 2. 数据库不存在
**错误**: `database "gochat" does not exist`

**解决方案**:
运行 SQL 命令创建数据库：
```sql
CREATE DATABASE gochat;
```

### 3. 端口已被占用
**错误**: `bind: address already in use`

**解决方案**:
- 修改服务器端口（在 main.go 中）
- 或者停止占用 8080 端口的其他程序

## 快速测试（无需 PostgreSQL）

如果你想快速测试代码结构，可以暂时注释掉数据库相关代码，但这样将无法使用完整功能。

## 生产环境部署

在生产环境中：
1. 使用强密码
2. 修改 Session 密钥
3. 启用 HTTPS
4. 配置防火墙规则
5. 使用环境变量管理敏感信息
