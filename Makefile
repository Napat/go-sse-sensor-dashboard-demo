# ตัวแปรสำหรับการตั้งค่า
BINARY_NAME=go-sse-server
BUILD_DIR=build
DOCKER_UAT_NAME=sensor-dashboard-uat
DOCKER_PROD_NAME=sensor-dashboard-prod
DOCKER_DEV_NAME=sensor-dashboard-dev

.PHONY: all
all: run

# รัน server แบบ development ด้วย hot-reload (Air)
.PHONY: run
run:
	@echo "Starting server with hot-reload in development mode t http://localhost:8080"
	@cd backend && APP_ENV=dev air -c .air.toml

# รัน server ในโหมด development บน container พร้อม hot-reload
.PHONY: dev
dev:
	@echo "Starting development environment in Docker container with hot-reload..."
	@docker compose -f docker-compose.dev.yml up -d app-dev
	@echo "Development environment is running at http://localhost:8081"
	@echo "Code changes will automatically reload"

# รัน server ในโหมด UAT บน container
.PHONY: uat
uat:
	@echo "Starting UAT environment in Docker container..."
	@make docker-run-uat

# รัน server ในโหมด Production บน container
.PHONY: prod
prod:
	@echo "Starting Production environment in Docker container..."
	@make docker-run-prod

# รัน tests
.PHONY: test
test:
	@echo "Running tests with coverage..."
	@cd backend && go test ./... -v -cover

# รัน tests พร้อมแสดงค่า coverage
.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	@cd backend && go test ./... -coverprofile=coverage.out && go tool cover -html=coverage.out

# สร้างไฟล์ binary สำหรับ production
.PHONY: build
build:
	@echo "Building backend for production..."
	@mkdir -p $(BUILD_DIR)/backend
	@cd backend && go build -o ../$(BUILD_DIR)/$(BINARY_NAME) main.go
	@echo "Copying frontend files..."
	@mkdir -p $(BUILD_DIR)/frontend
	@cp -r frontend/static $(BUILD_DIR)/frontend/
	@mkdir -p $(BUILD_DIR)/configs/backend
	@cp configs/backend/.env.prod $(BUILD_DIR)/configs/backend/.env.prod
	@echo "Build completed: $(BUILD_DIR)/$(BINARY_NAME)"

# เริ่มใช้งาน server ที่ build แล้ว
.PHONY: start
start: build
	@echo "Starting server from build in production mode..."
	@cd $(BUILD_DIR) && APP_ENV=prod ./$(BINARY_NAME)

# ทำความสะอาด build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build directory and artifacts..."
	@rm -rf $(BUILD_DIR)
	@rm -f backend/coverage.out

# ติดตั้งเครื่องมือสำหรับ development
.PHONY: install-dev-tools
install-dev-tools:
	@echo "Installing development tools..."
	@go install github.com/joho/godotenv@latest
	@go install github.com/cosmtrek/air@latest
	@echo "Tools installed successfully"

# ========================================
# Docker commands
# ========================================

# Build Docker images ทั้ง UAT และ Production
.PHONY: docker-build
docker-build:
	@echo "Building Docker images for all environments..."
	@docker compose -f docker-compose.dev.yml build app-dev
	@docker compose -f docker-compose.uat.yml build app-uat
	@docker compose -f docker-compose.prod.yml build app-prod

# รัน container สำหรับ UAT environment
.PHONY: docker-run-uat
docker-run-uat:
	@echo "Building and starting UAT environment container..."
	@docker compose -f docker-compose.uat.yml build app-uat
	@docker compose -f docker-compose.uat.yml up -d app-uat
	@echo "UAT environment is running at http://localhost:8082"

# รัน container สำหรับ Production environment
.PHONY: docker-run-prod
docker-run-prod:
	@echo "Building and starting Production environment container..."
	@docker compose -f docker-compose.prod.yml build
	@docker compose -f docker-compose.prod.yml up -d
	@echo "Production environment is running at http://localhost:8083"

# รัน container ทั้งหมด
.PHONY: docker-run-all
docker-run-all:
	@echo "Building and starting all containers..."
	@docker compose -f docker-compose.dev.yml build app-dev
	@docker compose -f docker-compose.dev.yml up -d app-dev
	@docker compose -f docker-compose.uat.yml build app-uat
	@docker compose -f docker-compose.uat.yml up -d app-uat
	@docker compose -f docker-compose.prod.yml build app-prod
	@docker compose -f docker-compose.prod.yml up -d app-prod
	@echo "Development environment is running at http://localhost:8081"
	@echo "UAT environment is running at http://localhost:8082"
	@echo "Production environment is running at http://localhost:8083"

# หยุด containers
.PHONY: docker-stop
docker-stop:
	@echo "Stopping all containers..."
	@docker compose -f docker-compose.dev.yml down
	@docker compose -f docker-compose.uat.yml down
	@docker compose -f docker-compose.prod.yml down

# ล้าง Docker images และ containers
.PHONY: docker-clean
docker-clean:
	@echo "Cleaning Docker resources..."
	@docker compose -f docker-compose.dev.yml down --rmi all --volumes --remove-orphans
	@docker compose -f docker-compose.uat.yml down --rmi all --volumes --remove-orphans
	@docker compose -f docker-compose.prod.yml down --rmi all --volumes --remove-orphans

# แสดงความช่วยเหลือ
.PHONY: help
help:
	@echo "Available commands:"
	@echo "  make run           - Start the server with hot-reload in development mode"
	@echo "  make dev           - Start the development environment in Docker container with hot-reload"
	@echo "  make uat           - Start the UAT environment in Docker container"
	@echo "  make prod          - Start the Production environment in Docker container"
	@echo "  make test          - Run all tests"
	@echo "  make test-coverage - Run tests with coverage report"
	@echo "  make build         - Build the application for production"
	@echo "  make start         - Build and start the production server locally"
	@echo "  make clean         - Remove build artifacts"
	@echo "  make install-dev-tools - Install development tools"
	@echo
	@echo "Docker commands:"
	@echo "  make docker-build  - Build Docker images for all environments"
	@echo "  make docker-run-uat - Run UAT environment in Docker (alias: make uat)"
	@echo "  make docker-run-prod - Run Production environment in Docker (alias: make prod)"
	@echo "  make docker-run-all - Run all Docker environments"
	@echo "  make docker-stop   - Stop all Docker containers"
	@echo "  make docker-clean  - Remove Docker images and containers"