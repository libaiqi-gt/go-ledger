# Docker Desktop 部署指南 (含热重载)

本文档详细记录了将 Go Ledger 项目部署到 Docker Desktop 的完整步骤，以及如何配置 Air 热重载。

## 1. 环境准备

- **安装 Docker Desktop**: 确保已安装并启动 Docker Desktop (Windows/Mac)。
- **开启 Kubernetes (可选)**: 本指南主要基于 Docker Compose，不需要 K8s。
- **项目结构**: 确保根目录下有以下关键文件：
    - `Dockerfile` (定义应用镜像构建)
    - `docker-compose.yml` (定义服务编排)
    - `.air.toml` (Air 热重载配置)

## 2. 详细部署步骤

### 步骤 1: 编写 Dockerfile

在根目录创建 `Dockerfile`，使用 Go 1.25 镜像并安装 Air 工具。

```dockerfile
# 使用官方 Golang 镜像作为构建环境
FROM golang:1.25

# 设置工作目录
WORKDIR /app

# 设置 Go 代理以加速下载
ENV GOPROXY=https://goproxy.cn,direct

# 安装 air 工具用于热重载
RUN go install github.com/air-verse/air@latest

# 复制依赖文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 暴露端口
EXPOSE 8080

# 使用 air 启动应用
CMD ["air"]
```

### 步骤 2: 编写 docker-compose.yml

在根目录创建 `docker-compose.yml`，编排 App 和 MySQL。
**注意**: 为了避免与本机 MySQL 端口冲突，我们将容器的 3306 映射到了宿主机的 **3307**。

```yaml
services:
  app:
    build: .
    container_name: go-ledger-app
    ports:
      - "8080:8080"
    volumes:
      - .:/app  # 挂载当前目录，实现代码同步
    environment:
      - DATABASE_HOST=db
      - DATABASE_PORT=3306
      - DATABASE_USER=root
      - DATABASE_PASSWORD=root_password
      - DATABASE_DBNAME=ledger_db
      - JWT_SECRET=docker_secret_key
    depends_on:
      - db
    restart: unless-stopped

  db:
    image: mysql:8.0
    container_name: go-ledger-db
    ports:
      - "3307:3306" # 映射到 3307 避免本地 3306 冲突
    environment:
      - MYSQL_ROOT_PASSWORD=root_password
      - MYSQL_DATABASE=ledger_db
    volumes:
      - mysql_data:/var/lib/mysql
    restart: unless-stopped

volumes:
  mysql_data:
```

### 步骤 3: 配置 Air 热重载

在根目录创建 `.air.toml`，并**务必开启 Poll 模式**。

```toml
root = "."
tmp_dir = "tmp"

[build]
  cmd = "go build -o ./tmp/main ."
  bin = "./tmp/main"
  poll = true          # 关键：Docker 下必须开启轮询
  poll_interval = 500  # 轮询间隔 500ms
  include_ext = ["go", "tpl", "tmpl", "html", "yaml"]
  exclude_dir = ["assets", "tmp", "vendor", "testdata"]

[log]
  time = false

[misc]
  clean_on_exit = true
```

### 步骤 4: 修改代码支持环境变量

确保 `config/database.go` 中启用了 Viper 的环境变量自动注入，以便读取 Docker Compose 传入的环境变量。

```go
func InitConfig() {
    // ... 原有代码 ...
    
    // 开启环境变量支持，将 . 替换为 _ (如 database.host -> DATABASE_HOST)
    viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
    viper.AutomaticEnv()
}
```

### 步骤 5: 启动服务

在项目根目录打开终端 (PowerShell 或 CMD)，执行：

```powershell
# 启动并后台运行
docker-compose up -d

# 如果修改了 Dockerfile，需要重新构建
docker-compose up -d --build
```

### 步骤 6: 验证

1. **查看运行日志**:
   ```powershell
   docker-compose logs -f app
   ```
   看到 `running...` 表示启动成功。
2. **测试热重载**:
   修改 `main.go` 中的打印语句并保存，观察日志，应能看到自动重新构建。
3. **访问接口**:
   访问 `http://localhost:8080/v1/entries` (需先获取 Token)。
4. **连接数据库**:
   使用数据库工具连接 `localhost:3307`，用户 `root`，密码 `root_password`。

---

## 3. 常见问题与解决方案 (Troubleshooting)

### Q1: `bind: Only one usage of each socket address is normally permitted`
- **原因**: 本机已经运行了 MySQL，占用了 3306 端口。
- **解决**: 在 `docker-compose.yml` 中修改 db 服务的端口映射，例如 `"3307:3306"`。

### Q2: `github.com/air-verse/air@v1.63.4 requires go >= 1.25`
- **原因**: `Dockerfile` 中的基础镜像版本过低 (如 golang:1.24)，而新版 Air 需要 Go 1.25+。
- **解决**: 将 `Dockerfile` 第一行改为 `FROM golang:1.25`，然后运行 `docker-compose up -d --build`。

### Q3: 修改代码后没有触发热重载
- **原因**: Windows Docker Desktop 的文件挂载机制限制，无法将文件变更事件实时通知给 Linux 容器。
- **解决**: 修改 `.air.toml`，设置 `poll = true` 开启轮询模式。

### Q4: 数据库连接失败 `dial tcp: lookup db on ...: no such host`
- **原因**: Go 代码中配置的 `host` 是 `127.0.0.1`，而在 Docker 网络中应使用服务名 `db`。
- **解决**: 确保 `docker-compose.yml` 中配置了环境变量 `DATABASE_HOST=db`，且 Go 代码中启用了 `viper.AutomaticEnv()`。

### Q5: 第一次启动很慢
- **原因**: 需要下载 Go 基础镜像和依赖包。
- **解决**: 确保 `Dockerfile` 中配置了 `ENV GOPROXY=https://goproxy.cn,direct` 国内代理。
