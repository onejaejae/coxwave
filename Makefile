.PHONY: run docker-up docker-down migration-up

# Run the application
run:
	go run ./cmd/main.go

# Start Docker containers
docker-up:
	docker-compose -f docker-compose.local.yml up -d

# Stop Docker containers
docker-down:
	docker-compose -f docker-compose.local.yml down 

docker-test-up:
	docker-compose -f docker-compose.test.yml up -d

docker-test-down:
	docker-compose -f docker-compose.test.yml down



# Run migration up
migration-up:
	go run ./internal/infrastructure/mysql/migration/migration_up.go
