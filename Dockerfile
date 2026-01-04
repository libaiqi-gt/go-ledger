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

# 复制源代码 (虽然 compose 会挂载，但为了构建缓存和生产环境建议保留)
COPY . .

# 暴露端口
EXPOSE 8080

# 使用 air 启动应用
CMD ["air"]
