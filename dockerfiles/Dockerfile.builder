# Multi-stage builder image for Go and Node.js development
FROM alpine:3.22 AS builder

LABEL maintainer="77kymo <opensource@kymo.cn>"
LABEL description="Builder image with Go, Node.js, and protobuf tools"
LABEL version="1.0"

USER root

# Set working directory
WORKDIR /app

# Create log directory for build steps
RUN mkdir -p /var/log/build

# Configure Alpine package manager with Aliyun mirror
RUN echo "Step 1: Configuring package repositories" > /var/log/build/build.log && \
    sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories && \
    echo "$(date): Package repositories configured" >> /var/log/build/build.log

# Update package index and install base dependencies
RUN echo "Step 2: Installing base packages" >> /var/log/build/build.log && \
    apk update && \
    apk add --no-cache ca-certificates tzdata && \
    echo "$(date): Base packages installed" >> /var/log/build/build.log

# Set timezone
ENV TZ=Asia/Shanghai
RUN echo "Step 3: Setting timezone to ${TZ}" >> /var/log/build/build.log && \
    echo "$(date): Timezone configured" >> /var/log/build/build.log

# Install essential build tools
RUN echo "Step 4: Installing build tools" >> /var/log/build/build.log && \
    apk add --no-cache make git dateutils tar wget zip unzip curl && \
    echo "$(date): Build tools installed" >> /var/log/build/build.log

# Download and install Go 1.24.2
RUN echo "Step 5: Installing Go 1.24.2" >> /var/log/build/build.log && \
    wget -q https://go.dev/dl/go1.24.2.linux-amd64.tar.gz && \
    rm -rf /usr/local/go && \
    tar -C /usr/local -xzf go1.24.2.linux-amd64.tar.gz && \
    rm go1.24.2.linux-amd64.tar.gz && \
    ln -sf /usr/local/go/bin/go /usr/local/bin/go && \
    echo "$(date): Go 1.24.2 installed successfully" >> /var/log/build/build.log

# Configure Go environment variables
ENV GOPROXY=https://goproxy.cn/,direct
ENV GOOS=linux
ENV GOARCH=amd64
ENV CGO_ENABLED=0
ENV GO111MODULE=on
ENV PATH="/usr/local/go/bin:${PATH}"

# Verify Go installation
RUN echo "Step 6: Verifying Go installation" >> /var/log/build/build.log && \
    go version && \
    echo "$(date): Go version: $(go version)" >> /var/log/build/build.log

# Install protobuf and gRPC tools
RUN echo "Step 7: Installing protobuf and gRPC tools" >> /var/log/build/build.log && \
    go install github.com/bufbuild/buf/cmd/buf@v1.59.0 && \
    echo "$(date): buf v1.59.0 installed" >> /var/log/build/build.log && \
    go install github.com/favadi/protoc-go-inject-tag@v1.4.0 && \
    echo "$(date): protoc-go-inject-tag v1.4.0 installed" >> /var/log/build/build.log && \
    go install github.com/go-swagger/go-swagger/cmd/swagger@v0.33.1 && \
    echo "$(date): go-swagger v0.33.1 installed" >> /var/log/build/build.log && \
    go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.35.2 && \
    echo "$(date): protoc-gen-go v1.35.2 installed" >> /var/log/build/build.log && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5.1 && \
    echo "$(date): protoc-gen-go-grpc v1.5.1 installed" >> /var/log/build/build.log && \
    go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@v2.27.3 && \
    echo "$(date): protoc-gen-grpc-gateway v2.27.3 installed" >> /var/log/build/build.log && \
    go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@v2.27.3 && \
    echo "$(date): protoc-gen-openapiv2 v2.27.3 installed" >> /var/log/build/build.log

# Install Node.js ecosystem
RUN echo "Step 8: Installing Node.js ecosystem" >> /var/log/build/build.log && \
    apk add --no-cache nodejs npm && \
    npm install -g pnpm && \
    echo "$(date): Node.js ecosystem installed" >> /var/log/build/build.log

# Configure npm registry
RUN echo "Step 9: Configuring npm registry" >> /var/log/build/build.log && \
    npm config set registry https://registry.npmmirror.com && \
    echo "$(date): npm registry configured" >> /var/log/build/build.log

# Verify Node.js installations and log versions
RUN echo "Step 10: Verifying Node.js installations" >> /var/log/build/build.log && \
    echo "$(date): Node.js version: $(node --version)" >> /var/log/build/build.log && \
    echo "$(date): npm version: $(npm --version)" >> /var/log/build/build.log && \
    echo "$(date): pnpm version: $(pnpm --version)" >> /var/log/build/build.log && \
    node --version && \
    npm --version && \
    pnpm --version

# Final build summary
RUN echo "Step 11: Build completed successfully" >> /var/log/build/build.log && \
    echo "$(date): All tools installed and configured" >> /var/log/build/build.log && \
    echo "Build Summary:" >> /var/log/build/build.log && \
    echo "- Alpine Linux: $(cat /etc/alpine-release)" >> /var/log/build/build.log && \
    echo "- Go: $(go version | cut -d' ' -f3)" >> /var/log/build/build.log && \
    echo "- Node.js: $(node --version)" >> /var/log/build/build.log && \
    echo "- npm: $(npm --version)" >> /var/log/build/build.log && \
    echo "- pnpm: $(pnpm --version)" >> /var/log/build/build.log

# docker build -t 77kymo/builder:v1 -f Dockerfile.builder .
# docker build -t ccr.ccs.tencentyun.com/itqm-private/builder:v1 -f Dockerfile.builder .