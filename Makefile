#!/usr/bin/make -f

# Project paths
ROOT_PATH := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
BACKEND_PATH := $(ROOT_PATH)/backend
FRONTEND_PATH := $(ROOT_PATH)/frontend
DOCKERFILES_PATH := $(ROOT_PATH)/dockerfiles

# Version and build info
VERSION := $(shell cat $(ROOT_PATH)/VERSION 2>/dev/null || echo "v1.0.0")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME := $(shell date +%Y%m%d%H%M%S)

# Container registry
IMAGE_REGISTRY := your-container-registry

# Go build environment
GO_PROXY ?= https://goproxy.cn/
GOARCH := $(shell go env GOARCH)
GOOS := linux
CGO_ENABLED ?= 0
GO_BUILD_ENV ?= GOPROXY=${GO_PROXY} GOOS=${GOOS} GOARCH=${GOARCH} CGO_ENABLED=${CGO_ENABLED}
GO_VERSION := $(shell go version | awk '{print $$3}')

# Build flags
VERSION_PKG := qm-mcp-server/pkg/version
LDFLAGS := -X '${VERSION_PKG}.Version=${VERSION}' \
		-X '${VERSION_PKG}.BuildTime=${BUILD_TIME}' \
		-X '${VERSION_PKG}.Commit=${COMMIT}' \
		-X '${VERSION_PKG}.GoVersion=${GO_VERSION}'

# Default target
.PHONY: all
all: help

.PHONY: print
print:
	@echo "---------- Project Configuration ----------"
	@echo "ROOT_PATH: $(ROOT_PATH)"
	@echo "BACKEND_PATH: $(BACKEND_PATH)"
	@echo "FRONTEND_PATH: $(FRONTEND_PATH)"
	@echo "DOCKERFILES_PATH: $(DOCKERFILES_PATH)"
	@echo "VERSION: $(VERSION)"
	@echo "COMMIT: $(COMMIT)"
	@echo "BUILD_TIME: $(BUILD_TIME)"
	@echo "GO_VERSION: $(GO_VERSION)"
	@echo "IMAGE_REGISTRY: $(IMAGE_REGISTRY)"
	@echo "GO_BUILD_ENV: $(GO_BUILD_ENV)"
	@echo "-------------------------------------------"

# Backend build targets
define build_backend_service
	@echo "---------- Start Go build $(1) ----------"
	@echo "cd $(BACKEND_PATH) && $(GO_BUILD_ENV) go build -ldflags \"$(LDFLAGS)\" -o $(BACKEND_PATH)/bin/$(1) $(BACKEND_PATH)/cmd/$(1)/main.go"
	@cd $(BACKEND_PATH) && $(GO_BUILD_ENV) go build -ldflags "$(LDFLAGS)" -o $(BACKEND_PATH)/bin/$(1) $(BACKEND_PATH)/cmd/$(1)/main.go
	@echo "---------- End Go build $(1) ----------"
endef

.PHONY: build-backend-init
build-backend-init:
	$(call build_backend_service,init)

.PHONY: build-backend-market
build-backend-market:
	$(call build_backend_service,market)

.PHONY: build-backend-authz
build-backend-authz:
	$(call build_backend_service,authz)

.PHONY: build-backend-gateway
build-backend-gateway:
	$(call build_backend_service,gateway)

.PHONY: build-backend-all
build-backend-all: build-backend-init build-backend-market build-backend-authz build-backend-gateway

# Frontend build targets
.PHONY: build-frontend
build-frontend:
	@echo "---------- Start build frontend ----------"
	@echo "cd $(FRONTEND_PATH) && pnpm i && pnpm build"
	@cd $(FRONTEND_PATH) && pnpm i && pnpm build
	@echo "---------- End build frontend ----------"

# Docker build targets
define build_docker_image
	@echo "---------- Start Docker build $(1) ----------"
	@echo "cd $(ROOT_PATH) && docker build -t $(IMAGE_REGISTRY)/$(2):$(VERSION) -f $(DOCKERFILES_PATH)/Dockerfile.$(1) ."
	@cd $(ROOT_PATH) && docker build -t $(IMAGE_REGISTRY)/$(2):$(VERSION) -f $(DOCKERFILES_PATH)/Dockerfile.$(1) .
	@echo "---------- End Docker build $(1) ----------"
endef

.PHONY: docker-build-init
docker-build-init:
	$(call build_docker_image,init,mcp-init)

.PHONY: docker-build-market
docker-build-market:
	$(call build_docker_image,market,mcp-market)

.PHONY: docker-build-authz
docker-build-authz:
	$(call build_docker_image,authz,mcp-authz)

.PHONY: docker-build-gateway
docker-build-gateway:
	$(call build_docker_image,gateway,mcp-gateway)

.PHONY: docker-build-frontend
docker-build-frontend:
	$(call build_docker_image,frontend,mcp-web)

.PHONY: docker-build-backend
docker-build-backend:
	$(call build_docker_image,backend,mcp-backend)

.PHONY: docker-build-all
docker-build-all: docker-build-init docker-build-market docker-build-authz docker-build-gateway docker-build-frontend

# Docker push targets
define push_docker_image
	@echo "---------- Start Docker push $(1) ----------"
	@echo "docker push $(IMAGE_REGISTRY)/$(1):$(VERSION)"
	@docker push $(IMAGE_REGISTRY)/$(1):$(VERSION)
	@echo "---------- End Docker push $(1) ----------"
endef

.PHONY: docker-push-init
docker-push-init:
	$(call push_docker_image,mcp-init)

.PHONY: docker-push-market
docker-push-market:
	$(call push_docker_image,mcp-market)

.PHONY: docker-push-authz
docker-push-authz:
	$(call push_docker_image,mcp-authz)

.PHONY: docker-push-gateway
docker-push-gateway:
	$(call push_docker_image,mcp-gateway)

.PHONY: docker-push-frontend
docker-push-frontend:
	$(call push_docker_image,mcp-web)

.PHONY: docker-push-backend
docker-push-backend:
	$(call push_docker_image,mcp-backend)

.PHONY: docker-push-all
docker-push-all: docker-push-init docker-push-market docker-push-authz docker-push-gateway docker-push-frontend

# Combined build and push targets
.PHONY: build-and-push-backend
build-and-push-backend: build-backend-all docker-build-init docker-build-market docker-build-authz docker-build-gateway docker-push-init docker-push-market docker-push-authz docker-push-gateway

.PHONY: build-and-push-frontend
build-and-push-frontend: build-frontend docker-build-frontend docker-push-frontend

.PHONY: build-and-push-all
build-and-push-all: build-and-push-backend build-and-push-frontend

# Protocol buffer generation
.PHONY: proto-buf
proto-buf:
	@echo "---- Cleaning existing generated files ----"
	@rm -rf $(shell find $(BACKEND_PATH)/api -type f -name '*.go')
	@rm -rf $(shell find $(BACKEND_PATH)/api -type f -name '*.json')
	@echo "---- Generating protobuf files ----"
	@cd $(BACKEND_PATH)/api && buf --debug generate 
	@find $(BACKEND_PATH)/api -name "*.pb.go" -exec protoc-go-inject-tag -input={} \; || echo "No .pb.go files found for tag injection"
	@echo "---- Merging swagger files ----"
	@swagger mixin $(shell rm -rf $(BACKEND_PATH)/api/merged.swagger.json && find $(BACKEND_PATH)/api -name "*.json") -o $(BACKEND_PATH)/api/merged.swagger.json 2>/dev/null || true
	@ls -la $(BACKEND_PATH)/api/merged.swagger.json

# Development targets
.PHONY: dev-backend
dev-backend:
	@echo "Starting backend development environment..."
	@cd $(BACKEND_PATH) && go run cmd/gateway/main.go

.PHONY: dev-frontend
dev-frontend:
	@echo "Starting frontend development environment..."
	@cd $(FRONTEND_PATH) && pnpm dev

# Clean targets
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BACKEND_PATH)/bin/*
	@rm -rf $(FRONTEND_PATH)/dist
	@rm -rf $(FRONTEND_PATH)/node_modules/.cache

.PHONY: clean-docker
clean-docker:
	@echo "Cleaning Docker images..."
	@docker images | grep "$(IMAGE_REGISTRY)" | awk '{print $$3}' | xargs -r docker rmi -f

# Test targets
.PHONY: test-backend
test-backend:
	@echo "Running backend tests..."
	@cd $(BACKEND_PATH) && go test ./...

.PHONY: test-frontend
test-frontend:
	@echo "Running frontend tests..."
	@cd $(FRONTEND_PATH) && pnpm test

.PHONY: test-all
test-all: test-backend test-frontend

# Lint targets
.PHONY: lint-backend
lint-backend:
	@echo "Linting backend code..."
	@cd $(BACKEND_PATH) && golangci-lint run

.PHONY: lint-frontend
lint-frontend:
	@echo "Linting frontend code..."
	@cd $(FRONTEND_PATH) && pnpm lint

.PHONY: lint-all
lint-all: lint-backend lint-frontend

# Help target
.PHONY: help
help:
	@echo "Available targets:"
	@echo ""
	@echo "Build targets:"
	@echo "  build-backend-init         - Build init service binary"
	@echo "  build-backend-market       - Build market service binary"
	@echo "  build-backend-authz        - Build authz service binary"
	@echo "  build-backend-gateway      - Build gateway service binary"
	@echo "  build-backend-all          - Build all backend services"
	@echo "  build-frontend             - Build frontend application"
	@echo ""
	@echo "Docker targets:"
	@echo "  docker-build-init          - Build init Docker image"
	@echo "  docker-build-market        - Build market Docker image"
	@echo "  docker-build-authz         - Build authz Docker image"
	@echo "  docker-build-gateway       - Build gateway Docker image"
	@echo "  docker-build-frontend      - Build frontend Docker image"
	@echo "  docker-build-all           - Build all Docker images"
	@echo "  docker-push-*              - Push corresponding Docker images"
	@echo "  docker-push-all            - Push all Docker images"
	@echo ""
	@echo "Combined targets:"
	@echo "  build-and-push-backend     - Build and push all backend services"
	@echo "  build-and-push-frontend    - Build and push frontend"
	@echo "  build-and-push-all         - Build and push everything"
	@echo ""
	@echo "Development targets:"
	@echo "  dev-backend                - Start backend development server"
	@echo "  dev-frontend               - Start frontend development server"
	@echo ""
	@echo "Utility targets:"
	@echo "  proto-buf                  - Generate protobuf and swagger files"
	@echo "  clean                      - Clean build artifacts"
	@echo "  clean-docker               - Clean Docker images"
	@echo "  test-all                   - Run all tests"
	@echo "  lint-all                   - Run all linters"
	@echo "  print                      - Print configuration"
	@echo "  help                       - Show this help message"