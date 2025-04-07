# 使用官方 Go 镜像作为基础镜像
FROM golang:1.22-alpine AS builder

ENV GOPROXY=https://goproxy.cn,direct

# 设置工作目录
WORKDIR /app

# 复制 go.mod 和 go.sum 文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY .. .

# 构建可执行文件
RUN go build -o myapp ./cmd/serverCmd.go

# 使用轻量级 Alpine 镜像作为运行环境
FROM alpine:latest

# 设置工作目录
WORKDIR /app

# 从构建阶段复制可执行文件
COPY --from=builder /app/myapp .

# 给指定文件添加权限
RUN chmod +x /app/myapp
RUN mkdir /app/data

# 设置容器启动命令
CMD ["./myapp"]