.PHONY: all test build start clean

all: test start clean

test:
	@echo "Running tests in the service-app-test container..."
	@docker compose build service-app-test || { echo "Failed to build the service-app-test image"; exit 1; }
	@docker compose run --rm service-app-test || { echo "Tests failed"; exit 1; }
	@docker compose down --remove-orphans --volumes service-app-test

integration-test:
	@${MAKE} start
	bash ./tester/test.sh
	@${MAKE} clean

build:
	@echo "Building the application..."
	@docker compose build service-app || { echo "Failed to build the application image"; exit 1; }

start:
	@${MAKE} build
	@echo "Running the application using Docker Compose..."
	@docker compose up -d service-app || { echo "Failed to start Docker Compose"; exit 1; }

start-sirius:
	@${MAKE} build
	@echo "Running the application using Docker Compose, integrated with local Sirius..."
	@docker compose -f docker-compose.yml -f docker-compose.sirius.yml up -d || { echo "Failed to start Docker Compose"; exit 1; }

clean:
	@echo "Stopping and cleaning up Docker Compose resources..."
	@docker compose down --remove-orphans --volumes || { echo "Failed to clean up resources"; exit 1; }

setup-directories:
	mkdir -p -m 0777 test-results

go-lint: setup-directories
	docker compose run --rm go-lint

gosec: setup-directories
	docker compose run --rm gosec

scan:
	docker compose run --rm trivy image --format table 311462405659.dkr.ecr.eu-west-1.amazonaws.com/sirius/scanning/app:latest
