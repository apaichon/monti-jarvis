BINARY := monti-jarvis
RUN_DIR := .run
PID := $(RUN_DIR)/server.pid
LOG := $(RUN_DIR)/server.log
PORT ?= 8091
CUSTOMER_WEB_DIR := apps/customer-web
COMPOSE_FILE := infra/docker-compose.yml

.PHONY: help build run start stop restart status logs test \
	customer-web customer-dev clean \
	infra-check infra-up infra-down infra-init infra-destroy infra-reset up down

help:
	@printf "App:\n"
	@printf "  make up             destroy + init infra, then start server\n"
	@printf "  make down           stop server + destroy infra\n"
	@printf "  make start          build and start server in background (:$(PORT))\n"
	@printf "  make stop           stop background server\n"
	@printf "  make restart        stop then start server\n"
	@printf "  make status         process + /healthz\n"
	@printf "  make logs           tail server log\n"
	@printf "  make run            foreground server\n"
	@printf "  make build          build customer-web + Go binary\n"
	@printf "  make customer-web   build Svelte portal only\n"
	@printf "  make customer-dev   vite dev on :5173 (proxies API)\n"
	@printf "  make test           go test ./...\n"
	@printf "Infra:\n"
	@printf "  make infra-reset    destroy then init all infra\n"
	@printf "  make infra-destroy  stop compose, drop DB, flush Redis, remove MinIO bucket\n"
	@printf "  make infra-init     create DB schema/tables and MinIO bucket\n"
	@printf "  make infra-up       docker compose up (NATS, LiveKit) + infra-init\n"
	@printf "  make infra-down     docker compose down (NATS, LiveKit)\n"
	@printf "  make infra-check    health check all services\n"

customer-web:
	@cd $(CUSTOMER_WEB_DIR) && npm install && npm run build

customer-dev:
	@cd $(CUSTOMER_WEB_DIR) && npm install && npm run dev

build: customer-web
	go build -o $(BINARY) ./cmd/server

run:
	PORT=$(PORT) go run ./cmd/server

$(RUN_DIR):
	@mkdir -p $(RUN_DIR)

start: build | $(RUN_DIR)
	@if [ -f $(PID) ] && kill -0 $$(cat $(PID)) 2>/dev/null; then \
		printf "already running (pid=%s)\n" $$(cat $(PID)); exit 1; \
	fi
	@PORT=$(PORT) nohup ./$(BINARY) >$(LOG) 2>&1 & echo $$! >$(PID)
	@sleep 0.6
	@if kill -0 $$(cat $(PID)) 2>/dev/null; then \
		printf "started pid=%s  url=http://localhost:$(PORT)  log=%s\n" $$(cat $(PID)) $(LOG); \
	else \
		printf "failed to start; last log lines:\n"; tail -n 40 $(LOG); rm -f $(PID); exit 1; \
	fi

stop:
	@if [ ! -f $(PID) ]; then printf "server not running\n"; exit 0; fi
	@PID_VALUE=$$(cat $(PID)); \
	if kill -0 $$PID_VALUE 2>/dev/null; then \
		kill $$PID_VALUE; \
		for i in 1 2 3 4 5; do kill -0 $$PID_VALUE 2>/dev/null || break; sleep 0.2; done; \
		kill -0 $$PID_VALUE 2>/dev/null && kill -9 $$PID_VALUE || true; \
		printf "stopped server pid=%s\n" $$PID_VALUE; \
	fi; \
	rm -f $(PID)

restart: stop start

status:
	@if [ -f $(PID) ] && kill -0 $$(cat $(PID)) 2>/dev/null; then \
		printf "server running pid=%s\n" $$(cat $(PID)); \
	else \
		printf "server not running\n"; \
	fi
	@curl -fsS http://localhost:$(PORT)/healthz 2>/dev/null && echo || echo "health: unreachable"

logs:
	@touch $(LOG)
	@tail -f $(LOG)

test:
	go test ./...

infra-check:
	@./scripts/infra-check.sh

infra-up:
	@./scripts/infra-up.sh

infra-down:
	@docker compose -f $(COMPOSE_FILE) down --remove-orphans 2>/dev/null || printf "compose down skipped\n"

infra-init:
	@./scripts/infra-init.sh

infra-destroy:
	@./scripts/infra-destroy.sh

infra-reset: infra-destroy infra-up

up: infra-reset start

down: stop infra-destroy

clean:
	rm -rf $(BINARY) $(RUN_DIR) $(CUSTOMER_WEB_DIR)/node_modules $(CUSTOMER_WEB_DIR)/build