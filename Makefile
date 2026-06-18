APP_NAME    := svr-admin
MODULE      := github.com/nostack-cn/svr-admin
VERSION     := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME  := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
COMMIT      := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS     := -s -w -X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.Commit=$(COMMIT)

# ---------- 开发 ----------

.PHONY: run
run: ## 本地启动（热重载需 air，否则直接 go run）
	go run .

.PHONY: dev
dev: ## 安装 air 并启动热重载
	@(command -v air > /dev/null || go install github.com/air-verse/air@latest)
	air

.PHONY: tidy
tidy: ## 整理依赖
	go mod tidy

.PHONY: deps
deps: ## 下载依赖
	go mod download

.PHONY: fmt
fmt: ## 格式化代码
	go fmt ./...
	gofmt -s -w .

.PHONY: lint
lint: ## 静态检查（需 golangci-lint）
	@golangci-lint run ./...

# ---------- 构建 ----------

.PHONY: build
build: ## 编译二进制到 bin/
	@mkdir -p bin
	go build -ldflags "$(LDFLAGS)" -o bin/server .

.PHONY: build-linux
build-linux: ## 交叉编译 Linux amd64
	@mkdir -p bin
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o bin/server-linux-amd64 .

# ---------- Docker ----------

.PHONY: docker-build
docker-build: ## 构建 Docker 镜像
	docker build -t $(APP_NAME):$(VERSION) -t $(APP_NAME):latest .

.PHONY: docker-run
docker-run: ## 运行 Docker 容器
	docker run --rm -p 8080:8080 --env-file .env $(APP_NAME):latest

# ---------- 测试 ----------

.PHONY: test
test: ## 运行测试
	go test -v -race ./...

.PHONY: coverage
coverage: ## 测试覆盖率
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# ---------- 清理 ----------

.PHONY: clean
clean: ## 清理构建产物
	rm -rf bin/ coverage.out coverage.html

# ---------- 帮助 ----------

.PHONY: help
help: ## 显示帮助信息
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-16s\033[0m %s\n", $$1, $$2}'
