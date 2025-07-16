.PHONY: build run test clean docker-up docker-down docker-test

# Variables
BINARY_NAME=transjakarta-fleet
DOCKER_COMPOSE_FILE=docker-compose.yml

# Build applications
build:
	@echo "🔨 Building applications..."
	go build -o bin/server cmd/server/main.go
	go build -o bin/publisher cmd/publisher/main.go
	go build -o bin/worker cmd/worker/main.go
	@echo "✅ Build completed"

# Run applications locally (requires infrastructure)
run:
	go run cmd/server/main.go

run-publisher:
	go run cmd/publisher/main.go

run-worker:
	go run cmd/worker/main.go

# Docker operations
docker-build:
	@echo "🔨 Building Docker images..."
	docker compose build

docker-up:
	@echo "🚀 Starting all services..."
	docker compose up -d

docker-down:
	@echo "🛑 Stopping all services..."
	docker compose down

docker-logs:
	docker compose logs -f

docker-restart:
	@echo "🔄 Restarting all services..."
	docker compose restart

# Infrastructure only (for local development)
docker-infra:
	@echo "🏗️ Starting infrastructure services only..."
	docker compose up postgres mosquitto rabbitmq -d

# Docker testing
docker-test:
	@echo "🧪 Starting Docker Integration Test"
	@chmod +x scripts/test_docker.sh
	@./scripts/test_docker.sh

# Phase 4 complete test
phase4-test: docker-test

# Production deployment
deploy-prod:
	@echo "🚀 Starting production deployment..."
	@chmod +x scripts/production_deploy.sh
	@./scripts/production_deploy.sh --confirm

# Development with Docker
dev-docker:
	@echo "🔥 Starting development mode with Docker..."
	docker compose -f docker-compose.yml -f docker/docker-compose.dev.yml up

# Postman testing
postman-test:
	@echo "📮 Running Postman API tests..."
	@chmod +x scripts/postman_test.sh
	@./scripts/postman_test.sh

# Complete testing suite
test-all: docker-test postman-test
	@echo "🎉 All tests completed!"

# Setup (one-command deployment)
setup:
	@echo "🚀 Starting complete setup..."
	@chmod +x scripts/setup.sh
	@./scripts/setup.sh

# Cleanup
clean:
	@echo "🧹 Cleaning up..."
	rm -rf bin/
	docker compose down -v
	docker system prune -f
	go clean
	@echo "✅ Cleanup completed"

# Database operations
db-connect:
	docker compose exec postgres psql -U postgres -d transjakarta_fleet

db-backup:
	@echo "💾 Creating database backup..."
	docker compose exec postgres pg_dump -U postgres transjakarta_fleet > backup-$(shell date +%Y%m%d_%H%M%S).sql
	@echo "✅ Backup created"

db-restore:
	@echo "⚠️ This will restore database from backup.sql"
	@read -p "Continue? (y/N): " confirm && [ "$confirm" = "y" ]
	docker compose exec -T postgres psql -U postgres transjakarta_fleet < backup.sql

# Monitoring
monitor:
	watch -n 2 'curl -s http://localhost:3000/api/v1/stats | jq .'

monitor-logs:
	docker compose logs -f --tail=50

monitor-health:
	watch -n 5 'curl -s http://localhost:3000/health | jq .'

# Scaling
scale-workers:
	docker compose up -d --scale worker=3

scale-servers:
	docker compose up -d --scale server=2

# Development helpers
dev-setup: docker-infra
	@echo "🔥 Development environment ready!"
	@echo "Run in separate terminals:"
	@echo "  make run         # Start server"
	@echo "  make run-publisher # Start publisher"
	@echo "  make run-worker  # Start worker"

# Show help
help:
	@echo "🚀 TransJakarta Fleet Management - Available Commands"
	@echo "=================================================="
	@echo ""
	@echo "🏗️  Setup & Deployment:"
	@echo "  setup            - Complete one-command setup"
	@echo "  docker-up        - Start all services with Docker"
	@echo "  docker-infra     - Start infrastructure only"
	@echo "  deploy-prod      - Production deployment"
	@echo ""
	@echo "🔨 Development:"
	@echo "  build            - Build all applications"
	@echo "  run              - Run server locally"
	@echo "  run-publisher    - Run publisher locally"
	@echo "  run-worker       - Run worker locally"
	@echo "  dev-docker       - Development mode with hot reload"
	@echo "  dev-setup        - Setup for local development"
	@echo ""
	@echo "🧪 Testing:"
	@echo "  docker-test      - Complete Docker integration test"
	@echo "  postman-test     - API testing with Postman"
	@echo "  test-all         - Run all test suites"
	@echo ""
	@echo "🐳 Docker Management:"
	@echo "  docker-build     - Build Docker images"
	@echo "  docker-down      - Stop all services"
	@echo "  docker-logs      - View all logs"
	@echo "  docker-restart   - Restart all services"
	@echo ""
	@echo "📊 Monitoring:"
	@echo "  monitor          - Real-time system statistics"
	@echo "  monitor-logs     - Live log monitoring"
	@echo "  monitor-health   - Health check monitoring"
	@echo ""
	@echo "🗄️  Database:"
	@echo "  db-connect       - Connect to PostgreSQL"
	@echo "  db-backup        - Create database backup"
	@echo "  db-restore       - Restore from backup"
	@echo ""
	@echo "⚖️  Scaling:"
	@echo "  scale-workers    - Scale worker services to 3"
	@echo "  scale-servers    - Scale server services to 2"
	@echo ""
	@echo "🧹 Maintenance:"
	@echo "  clean            - Clean up all artifacts"
	@echo ""
	@echo "🔍 Quick Start:"
	@echo "  make setup       # One command to get everything running"
	@echo "  make docker-test # Verify everything works"

# Default target
.DEFAULT_GOAL := help