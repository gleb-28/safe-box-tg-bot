COMPOSE_FILE=deploy/docker-compose.yml

.PHONY: build down rebuild logs tidy run deploy

# Docker commands
build:
	@echo "Building Docker images..."
	docker-compose -f $(COMPOSE_FILE) up -d
down:
	@echo "Stopping Docker containers and cleaning up volumes and images..."
	docker-compose -f $(COMPOSE_FILE) down -v --rmi all
rebuild:
	@echo "Rebuilding and restarting Docker containers..."
	docker-compose -f $(COMPOSE_FILE) up -d --build
logs:
	@echo "Viewing Docker container logs (press Ctrl+C to exit)..."
	docker-compose -f $(COMPOSE_FILE) logs -f
deploy:
	./deploy/deploy.sh


# Go commands
tidy:
	@echo "Tidying Go modules..."
	go mod tidy

run:
	@echo "Running Go application locally..."
	go run cmd/bot/main.go

test:
	go test ./... -v
