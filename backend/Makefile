HOME_PATH := $(HOME)
PROJ_PATH := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
API_PB_PATH := $(PROJ_PATH)/api
SWAG_BIN := $(shell which swagger)

# 获取用户目录 和 项目路径
BINARY_NAME=$(PROJ_PATH)/bin/qm-mcp-server
VERSION=$(shell cat $(PROJ_PATH)/VERSION)
TAG=$(VERSION)
DOCKER_IMAGE=ccr.ccs.tencentyun.com/itqm-private/qm-mcp-server
DOCKER_COMPOSE_DIR=$(PROJ_PATH)/deploy/docker-compose
ENV_TRANSFER:=HOME_PATH=$(HOME_PATH) PROJ_PATH=$(PROJ_PATH) VERSION=$(VERSION) TAG=$(TAG) DOCKER_IMAGE=$(DOCKER_IMAGE) BINARY_NAME=$(BINARY_NAME) DOCKER_COMPOSE_DIR=$(DOCKER_COMPOSE_DIR)


QM_IMAGE_REGISTRY=ccr.ccs.tencentyun.com/itqm-private

COMMIT=$(shell git rev-parse --short HEAD)

BUILD_TIME=$(shell date +%Y%m%d%H%M%S)

LinuxGoos := linux
DarwinGoos := darwin
WindowsGoos := windows

GO_MOD_PATH := $(subst \,/,$(GOPATH))
GO_PROXY ?= https://goproxy.cn/
GOARCH := $(shell go env GOARCH)
GOOS := ${LinuxGoos}
GOPATH := $(shell go env GOPATH)
CGO_ENABLED ?= 0
GO_BUILD_ENV ?= GOPROXY=${GO_PROXY} GOOS=${GOOS} GOARCH=${GOARCH}  CGO_ENABLED=${CGO_ENABLED}
GO_VERSION := $(shell go version | awk '{print $$3}')
VERSION_PKG := qm-mcp-server/pkg/version
LDFLAGS := -X '${VERSION_PKG}.Version=${VERSION}' \
		-X '${VERSION_PKG}.BuildTime=${BUILD_TIME}' \
		-X '${VERSION_PKG}.Commit=${COMMIT}' \
		-X '${VERSION_PKG}.GoVersion=${GO_VERSION}'

# 默认目标
.PHONY: all
all: print

.PHONY: print
print:
	@echo "---------- Start print --------------"
	@echo "HOME_PATH: $(HOME_PATH)"
	@echo "PROJ_PATH: $(PROJ_PATH)"
	@echo "VERSION: $(VERSION)"
	@echo "GO_VERSION: $(GO_VERSION)"
	@echo "BUILD_TIME: $(BUILD_TIME)"
	@echo "COMMIT: $(COMMIT)"
	@echo "TAG: $(TAG)"
	@echo "DOCKER_IMAGE: $(DOCKER_IMAGE)"
	@echo "BINARY_NAME: $(BINARY_NAME)"
	@echo "DOCKER_COMPOSE_DIR: $(DOCKER_COMPOSE_DIR)"
	@echo "ENV_TRANSFER: $(ENV_TRANSFER)"
	@echo "GO_BUILD_ENV: $(GO_BUILD_ENV)"
	@echo "---------- End print --------------"

.PHONY: local-build-push
local-build-push: go-build-init go-build-market go-build-authz go-build-gateway docker-build-init docker-build-market docker-build-authz docker-build-gateway docker-push-init docker-push-market docker-push-authz docker-push-gateway


# 构建目标模板函数
define build_service
	@echo "---------- Start Go build $(1) ----------"
	@echo "cd $(PROJ_PATH) && $(GO_BUILD_ENV) go build -ldflags \"$(LDFLAGS)\" -o $(PROJ_PATH)/bin/$(1) $(PROJ_PATH)/cmd/$(1)/main.go"
	@cd $(PROJ_PATH) && $(GO_BUILD_ENV) go build -ldflags "$(LDFLAGS)" -o $(PROJ_PATH)/bin/$(1) $(PROJ_PATH)/cmd/$(1)/main.go
	@echo "---------- End Go build $(1) ----------"
endef

.PHONY: go-build-init
go-build-init:
	$(call build_service,init)

.PHONY: go-build-market
go-build-market:
	$(call build_service,market)

.PHONY: go-build-authz
go-build-authz:
	$(call build_service,authz)

.PHONY: go-build-gateway
go-build-gateway:
	$(call build_service,gateway)

.PHONY: docker-build-init
docker-build-init:
	@echo "---------- Start Docker build init ----------"
	@echo "cd $(PROJ_PATH) && docker build -t $(QM_IMAGE_REGISTRY)/mcp-init:$(VERSION) -f Dockerfile.init ."
	@cd $(PROJ_PATH) && docker build -t $(QM_IMAGE_REGISTRY)/mcp-init:$(VERSION) -f Dockerfile.init .
	@echo "---------- End Docker build init ----------"

.PHONY: docker-build-market
docker-build-market:
	@echo "---------- Start Docker build market ----------"
	@echo "cd $(PROJ_PATH) && docker build -t $(QM_IMAGE_REGISTRY)/mcp-market:$(VERSION) -f Dockerfile.market ."
	@cd $(PROJ_PATH) && docker build -t $(QM_IMAGE_REGISTRY)/mcp-market:$(VERSION) -f Dockerfile.market .
	@echo "---------- End Docker build market ----------"

.PHONY: docker-build-authz
docker-build-authz:
	@echo "---------- Start Docker build authz ----------"
	@echo "cd $(PROJ_PATH) && docker build -t $(QM_IMAGE_REGISTRY)/mcp-authz:$(VERSION) -f Dockerfile.authz ."
	@cd $(PROJ_PATH) && docker build -t $(QM_IMAGE_REGISTRY)/mcp-authz:$(VERSION) -f Dockerfile.authz .
	@echo "---------- End Docker build authz ----------"

.PHONY: docker-build-gateway
docker-build-gateway:
	@echo "---------- Start Docker build gateway ----------"
	@echo "cd $(PROJ_PATH) && docker build -t $(QM_IMAGE_REGISTRY)/mcp-gateway:$(VERSION) -f Dockerfile.gateway ."
	@cd $(PROJ_PATH) && docker build -t $(QM_IMAGE_REGISTRY)/mcp-gateway:$(VERSION) -f Dockerfile.gateway .
	@echo "---------- End Docker build gateway ----------"

.PHONY: docker-push-init
docker-push-init:
	@echo "---------- Start Docker push init ----------"
	@echo "docker push $(QM_IMAGE_REGISTRY)/mcp-init:$(VERSION)"
	@docker push $(QM_IMAGE_REGISTRY)/mcp-init:$(VERSION)
	@echo "---------- End Docker push init ----------"

.PHONY: docker-push-market
docker-push-market:
	@echo "---------- Start Docker push market ----------"
	@echo "docker push $(QM_IMAGE_REGISTRY)/mcp-market:$(VERSION)"
	@docker push $(QM_IMAGE_REGISTRY)/mcp-market:$(VERSION)
	@echo "---------- End Docker push market ----------"

.PHONY: docker-push-authz
docker-push-authz:
	@echo "---------- Start Docker push authz ----------"
	@echo "docker push $(QM_IMAGE_REGISTRY)/mcp-authz:$(VERSION)"
	@docker push $(QM_IMAGE_REGISTRY)/mcp-authz:$(VERSION)
	@echo "---------- End Docker push authz ----------"

.PHONY: docker-push-gateway
docker-push-gateway:
	@echo "---------- Start Docker push gateway ----------"
	@echo "docker push $(QM_IMAGE_REGISTRY)/mcp-gateway:$(VERSION)"
	@docker push $(QM_IMAGE_REGISTRY)/mcp-gateway:$(VERSION)
	@echo "---------- End Docker push gateway ----------"


.PHONY: proto-buf
proto-buf:
	@echo "---- 正在执行清理操作 ----"
	@rm -rf $(shell find $(API_PB_PATH) -type f -name '*.go')
	@rm -rf $(shell find $(API_PB_PATH) -type f -name '*.json')
	@echo "---- 正在执行 Buf 操作 (格式化、检查、更新依赖、生成代码) ----"
	@cd $(API_PB_PATH) && buf --debug generate 
	@find $(API_PB_PATH) -name "*.pb.go" -exec protoc-go-inject-tag -input={} \; || echo "No .pb.go files found for tag injection"
	@echo "---- 正在执行 swagger mixin 操作 (合并 swagger.json 文件) ----"
	@echo "$(SWAG_BIN) mixin $(shell rm -rf $(API_PB_PATH)/merged.swagger.json && find $(API_PB_PATH) -name "*.json") -o $(API_PB_PATH)/merged.swagger.json" 2>/dev/null || true
	@$(SWAG_BIN) mixin $(shell rm -rf $(API_PB_PATH)/merged.swagger.json && find $(API_PB_PATH) -name "*.json") -o $(API_PB_PATH)/merged.swagger.json 2>/dev/null || true
	@ls -la $(API_PB_PATH)/merged.swagger.json

# K8s 部署
.PHONY: k8s-deploy
k8s-deploy:
	@echo "---------- Start K8s deployment ----------"
	@echo "cd $(PROJ_PATH)/../deploy/scripts && ./install-k3s.sh"
	@cd $(PROJ_PATH)/../deploy/scripts && ./install-k3s.sh
	@echo "cd $(PROJ_PATH)/../deploy/scripts && ./install-mcp.sh"
	@cd $(PROJ_PATH)/../deploy/scripts && ./install-mcp.sh
	@echo "---------- End K8s deployment ----------"

# K8s 卸载
.PHONY: k8s-uninstall
k8s-uninstall:
	@echo "---------- Start K8s uninstall ----------"
	@echo "cd $(PROJ_PATH)/../deploy/scripts && ./uninstall-mcp.sh"
	@cd $(PROJ_PATH)/../deploy/scripts && ./uninstall-mcp.sh
	@echo "cd $(PROJ_PATH)/../deploy/scripts && ./uninstall-k3s.sh"
	@cd $(PROJ_PATH)/../deploy/scripts && ./uninstall-k3s.sh
	@echo "---------- End K8s uninstall ----------"

# 完整部署流程
.PHONY: deploy-all
deploy-all: go-build-market go-build-authz go-build-gateway go-build-init docker-build-market docker-build-authz docker-build-gateway docker-build-init docker-push-market docker-push-authz docker-push-gateway docker-push-init k8s-deploy
	@echo "---------- All components deployed successfully! ----------"

# 打印帮助信息
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  print                      - Print current configuration"
	@echo "  local-build-push           - Build and push locally (market, authz, gateway)"
	@echo "  go-build-market            - Build market binary"
	@echo "  go-build-authz            - Build authz binary"
	@echo "  go-build-gateway          - Build gateway binary"
	@echo "  docker-build-market       - Build market docker image"
	@echo "  docker-build-authz        - Build authz docker image"
	@echo "  docker-build-gateway      - Build gateway docker image"
	@echo "  docker-push-market        - Push market docker image"
	@echo "  docker-push-authz         - Push authz docker image"
	@echo "  docker-push-gateway       - Push gateway docker image"
	@echo "  go-build-init             - Build init binary"
	@echo "  docker-build-init         - Build init docker image"
	@echo "  docker-push-init          - Push init docker image"
	@echo "  k8s-deploy                - Deploy to Kubernetes"
	@echo "  k8s-uninstall             - Uninstall from Kubernetes"
	@echo "  deploy-all                - Deploy all components"
	@echo "  proto-buf                 - Generate protobuf and swagger files"
	@echo "  help                      - Print this help message"



