# 构建阶段
FROM ccr.ccs.tencentyun.com/itqm-private/golang:1.24.2 AS builder

USER root

# 设置 apk aliyun 源
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

# 设置传输 VERSION 入参
ARG VERSION
ENV VERSION=${VERSION}

# 设置工作目录
WORKDIR /app

# 复制源代码
COPY . /app

# 下载依赖
RUN cd /app && \
    go mod tidy && \
    go build -o /app/bin/qm-mcp-server ./cmd/main.go

# 运行阶段
FROM ccr.ccs.tencentyun.com/itqm-private/alpine:3.22

# 设置 apk aliyun 源
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

# 安装必要的运行时依赖
RUN apk add --no-cache ca-certificates tzdata

# 设置时区
ENV TZ=Asia/Shanghai

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/bin/qm-mcp-server .

RUN mkdir -p /app/configs

# 运行应用
CMD ["/app/qm-mcp-server"] 