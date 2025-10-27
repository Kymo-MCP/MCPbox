PROJ_PATH:=$(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))

VERSION=$(shell cat $(PROJ_PATH)/VERSION)

QM_IMAGE_REGISTRY=ccr.ccs.tencentyun.com/itqm-private

COMMIT_ID=$(shell git rev-parse --short HEAD)

.PHONY: print
print:
	@echo "---------- Start print --------------"
	@echo "PROJ_PATH: $(PROJ_PATH)"
	@echo "VERSION: $(VERSION)"
	@echo "---------- End print --------------"

.PHONY: local-build-push
local-build-push: pnpm-build docker-build-web docker-push-web

.PHONY: pnpm-build
pnpm-build:
	@echo "---------- Start build web ----------"
	@echo "cd $(PROJ_PATH) && pnpm i && pnpm build"
	@cd $(PROJ_PATH) && pnpm i && pnpm build
	@echo "---------- End build web ----------"

.PHONY: docker-build-web
docker-build-web:
	@echo "---------- Start Docker build web ----------"
	@echo "cd $(PROJ_PATH) && docker build -t $(QM_IMAGE_REGISTRY)/mcp-web:$(VERSION) -f Dockerfile ."
	@cd $(PROJ_PATH) && docker build -t $(QM_IMAGE_REGISTRY)/mcp-web:$(VERSION) -f Dockerfile .
	@echo "---------- End Docker build web ----------"

.PHONY: docker-push-web
docker-push-web:
	@echo "---------- Start Docker push web ----------"
	@echo "docker push $(QM_IMAGE_REGISTRY)/mcp-web:$(VERSION)"
	@docker push $(QM_IMAGE_REGISTRY)/mcp-web:$(VERSION)
	@echo "---------- End Docker push web ----------"



