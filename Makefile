.PHONY: all test clean start

all: test start clean

test:
	@echo "Running tests in the service-app-test container..."
	@docker-compose build service-app-test || { echo "Failed to build the service-app-test image"; exit 1; }
	@docker-compose run --rm service-app-test || { echo "Tests failed"; exit 1; }
	@docker-compose down --remove-orphans --volumes service-app-test

start:
	@echo "Building the application..."
	@docker-compose build || { echo "Failed to build the application image"; exit 1; }
	@echo "Running the application using Docker Compose..."
	@docker-compose up -d || { echo "Failed to start Docker Compose"; exit 1; }

clean:
	@echo "Stopping and cleaning up Docker Compose resources..."
	@docker-compose down --remove-orphans --volumes || { echo "Failed to clean up resources"; exit 1; }
