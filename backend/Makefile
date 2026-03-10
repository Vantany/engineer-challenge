APP_NAME=auth-service

ifneq (,$(wildcard ./.env.example))
    include .env.example
    export
endif


DB_DSN ?= postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)

.PHONY: gen-keys
gen-keys:
	@mkdir -p keys
	openssl genpkey -algorithm RSA -out keys/private.pem -pkeyopt rsa_keygen_bits:2048
	openssl rsa -pubout -in keys/private.pem -out keys/public.pem

.PHONY: run
run:
	go run ./cmd/server

.PHONY: build
build:
	go build -o $(APP_NAME) ./cmd/server

.PHONY: test-unit
test-unit:
	go test -v -race ./internal/...

.PHONY: test-integration
test-integration:
	go test -v -race ./tests/integration/...

.PHONY: test
test: test-unit test-integration


.PHONY: proto
proto:
	export PATH="$$PATH:$$(go env GOPATH)/bin" && easyp generate

.PHONY: up
up:
	docker-compose up -d --build

.PHONY: down
down:
	docker-compose down

.PHONY: logs
logs:
	docker-compose logs -f auth-service

.PHONY: migrate-up
migrate-up:
	goose -dir migrations postgres $(DB_DSN) up

.PHONY: migrate-down
migrate-down:
	goose -dir migrations postgres $(DB_DSN) down

.PHONY: migrate-status
migrate-status:
	goose -dir migrations postgres $(DB_DSN) status

.PHONY: helm-install
helm-install:
	@echo "Installing Helm chart..."
	helm install auth-service deploy/helm/auth-service

.PHONY: helm-upgrade
helm-upgrade:
	@echo "Upgrading Helm chart..."
	helm upgrade auth-service deploy/helm/auth-service

.PHONY: helm-uninstall
helm-uninstall:
	@echo "Uninstalling Helm chart..."
	helm uninstall auth-service

.PHONY: deploy
deploy:
	@echo "Deploying to local Kubernetes (requires running local cluster)..."
	docker-compose up -d postgres migrator
	docker build -t auth-service:latest .
	helm upgrade --install auth-service deploy/helm/auth-service
	@echo "Deployed! You can check pods with: kubectl get pods"

