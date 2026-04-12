.PHONY: help dev run build test lint clean migrate db-up db-down db-logs db-reset gen gen-permissions \
        api api-bundle api-lint api-gen api-perms api-front

# ── OpenAPI 生成链路 ─────────────────────────────────────────────────────────
#
# 单一入口：make api（bundle → lint → ogen → gen-permissions → 前端 ts）
# 分步入口：api-bundle / api-lint / api-gen / api-perms / api-front
#
# 真源：api/openapi/openapi.yaml（Step 4 拆 domain 后改为 openapi.root.yaml）
# 产物：api/openapi/dist/openapi.yaml（ogen / 前端 ts 的输入，机器生成勿手改）

SPEC_SRC  = api/openapi/openapi.root.yaml
SPEC_DIST = api/openapi/dist/openapi.yaml
REDOCLY   = npx --yes @redocly/cli

api: api-gen api-perms ## 全量生成（bundle→lint→ogen→权限种子）

api-bundle: ## bundle 多文件 spec → dist/openapi.yaml
	@echo "[api] bundling spec..."
	@$(REDOCLY) bundle $(SPEC_SRC) -o $(SPEC_DIST) --ext yaml

api-lint: api-bundle ## bundle 后 lint 规则检查
	@echo "[api] linting spec..."
	@$(REDOCLY) lint $(SPEC_DIST) --config api/openapi/redocly.yaml

api-gen: api-lint ## bundle→lint→ogen 生成 Go server
	@echo "[api] running ogen..."
	@go run github.com/ogen-go/ogen/cmd/ogen@latest \
	  --target api/gen --package gen --clean $(SPEC_DIST)

api-perms: api-bundle ## bundle→权限种子（不依赖 ogen）
	@echo "[api] generating permission seed..."
	@go run ./cmd/gen-permissions

api-front: api-bundle ## bundle→前端 TS 类型（需在 frontend/ 配 gen:api 脚本）
	@echo "[api] generating frontend types..."
	@cd ../frontend && npm run gen:api

# ── 常用命令与兼容别名 ─────────────────────────────────────────────────────

help: ## 显示帮助信息
	@echo "可用命令:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

dev: ## 启动开发环境（直接运行，不使用 Air 热重载）
	@echo "启动开发服务器..."
	@go run cmd/server/main.go

run: ## 直接运行（不使用热重载）
	@go run cmd/server/main.go

build: ## 编译项目
	@echo "编译项目..."
	@go build -o bin/server cmd/server/main.go

test: ## 运行测试
	@go test ./... -v

lint: ## 代码检查
	@golangci-lint run

clean: ## 清理构建文件
	@rm -rf bin/ tmp/
	@echo "清理完成"

migrate: ## 运行数据库迁移
	@echo "运行数据库迁移..."
	@go run cmd/migrate/main.go

gen: api-gen ## OpenAPI 生成 Go server/client 代码（api-gen 的简写别名）

gen-permissions: api-perms ## 权限种子生成（api-perms 的简写别名）

deps: ## 安装依赖
	@go mod download
	@go mod tidy
	@echo "依赖安装完成"

# Docker Compose 命令
db-up: ## 启动数据库服务（PostgreSQL + Redis）
	@echo "启动数据库服务..."
	@docker-compose up -d postgres redis
	@echo "等待数据库就绪..."
	@sleep 3
	@docker-compose ps

db-up-all: ## 启动所有服务（包括 ES 和 MinIO）
	@echo "启动所有服务..."
	@docker-compose up -d
	@echo "等待服务就绪..."
	@sleep 5
	@docker-compose ps

db-down: ## 停止数据库服务
	@docker-compose stop

db-down-all: ## 停止并删除所有容器
	@docker-compose down

db-reset: ## 重置数据库（删除所有数据，谨慎使用！）
	@echo "⚠️  警告：这将删除所有数据库数据！"
	@docker-compose down -v
	@docker-compose up -d postgres redis
	@echo "数据库已重置"

db-logs: ## 查看数据库日志
	@docker-compose logs -f postgres redis

db-status: ## 查看服务状态
	@docker-compose ps

db-shell: ## 进入 PostgreSQL 命令行
	@docker exec -it gg-postgres psql -U postgres -d gg_ecommerce

db-redis: ## 进入 Redis 命令行
	@docker exec -it gg-redis redis-cli
