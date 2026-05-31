COMPOSE := docker compose -f deploy/docker-compose.yml
MOCKGEN := go run go.uber.org/mock/mockgen

.PHONY: help infra-up infra-down run build test test-race mock deps fmt vet

help:
	@echo "make infra-up    启动 postgres + redis"
	@echo "make infra-down  停止并清理基础设施"
	@echo "make run         本地启动后端 (go run ./cmd/server)"
	@echo "make build       go build ./..."
	@echo "make test        go test ./... -cover"
	@echo "make test-race   go test ./... -race -cover (CI)"
	@echo "make mock        用 mockgen 重新生成 internal/mock"
	@echo "make deps        go mod download"

infra-up:
	$(COMPOSE) up -d postgres redis

infra-down:
	$(COMPOSE) down

# 启动完整栈（含 backend 镜像、可选 rabbitmq）
infra-up-full:
	$(COMPOSE) --profile full up -d

run:
	go run ./cmd/server

build:
	go build ./...

test:
	go test ./... -cover

# CI 用：竞态检测（依赖 cgo，会自动走外部链接）
test-race:
	go test ./... -race -cover

deps:
	go mod download

fmt:
	go fmt ./...

vet:
	go vet ./...

# 重新生成 mock（接口变更后执行）。
mock:
	$(MOCKGEN) -source=internal/integration/market/provider.go -destination=internal/mock/market_provider_mock.go -package=mock
	$(MOCKGEN) -source=internal/repository/watchlist.go -destination=internal/mock/watchlist_repository_mock.go -package=mock
	$(MOCKGEN) -source=internal/repository/market_quote.go -destination=internal/mock/market_quote_repository_mock.go -package=mock
	$(MOCKGEN) -source=internal/service/market.go -destination=internal/mock/market_service_mock.go -package=mock
