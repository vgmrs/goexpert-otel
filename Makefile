.PHONY: install
install:
	@cd service-a && go mod tidy
	@cd ..
	@cd service-b && go mod tidy
	@cd ..

.PHONY: up
up:
	@docker compose up -d --build

.PHONY: down
down:
	@docker compose down

.PHONY: logs
logs:
	@docker compose logs -f
