BINARY := monti-jarvis
RUN_DIR := .run
PID := $(RUN_DIR)/server.pid
LOG := $(RUN_DIR)/server.log
PORT ?= 8091

.PHONY: help build run start stop restart status logs test infra-check infra-init clean

help:
	@printf "Targets:\n"
	@printf "  make run          run the server in foreground\n"
	@printf "  make start        run the server in background on :$(PORT)\n"
	@printf "  make stop         stop the background server\n"
	@printf "  make status       show process and health\n"
	@printf "  make infra-check  check shared Postgres/Redis/MinIO\n"
	@printf "  make infra-init   create isolated DB/schema and MinIO bucket\n"
	@printf "  make test         run Go tests\n"

build:
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
	@if [ ! -f $(PID) ]; then printf "not running\n"; exit 0; fi
	@PID_VALUE=$$(cat $(PID)); \
	if kill -0 $$PID_VALUE 2>/dev/null; then \
		kill $$PID_VALUE; \
		for i in 1 2 3 4 5; do kill -0 $$PID_VALUE 2>/dev/null || break; sleep 0.2; done; \
		kill -0 $$PID_VALUE 2>/dev/null && kill -9 $$PID_VALUE || true; \
		printf "stopped pid=%s\n" $$PID_VALUE; \
	fi; \
	rm -f $(PID)

restart: stop start

status:
	@if [ -f $(PID) ] && kill -0 $$(cat $(PID)) 2>/dev/null; then \
		printf "running pid=%s\n" $$(cat $(PID)); \
	else \
		printf "not running\n"; \
	fi
	@curl -fsS http://localhost:$(PORT)/healthz 2>/dev/null && echo || echo "health: unreachable"

logs:
	@touch $(LOG)
	@tail -f $(LOG)

test:
	go test ./...

infra-check:
	@./scripts/infra-check.sh

infra-init:
	@./scripts/infra-init.sh

clean:
	rm -rf $(BINARY) $(RUN_DIR)