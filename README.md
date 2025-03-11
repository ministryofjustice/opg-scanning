# OPG Scanning

OPG Scanning is a Go-based microservice designed to perform scanning and analysis operations. It is containerized with Docker Compose for both standard and Sirius-integrated environments.

## Overview

The OPG Scanning service enables you to run tests, build the application, and perform security and lint analysis through a set of defined Makefile commands. The project is structured for rapid local development and integration testing with external dependencies such as Localstack and a Sirius mock service.

## Requirements

For local development and testing you will need:

- [Docker](https://www.docker.com/get-started)
- [Docker Compose](https://docs.docker.com/compose/install/)
- [Golang](https://golang.org) (if you plan on running tests outside of Docker)

## Local Development

The Makefile includes several targets to simplify common development tasks:

- **Build and Test**
  - **`make`**: Runs tests, starts the application, and then cleans up.
  - **`make test`**: Builds the `service-app-test` container and runs unit tests.
  - **`make integration-test`**: Starts the application, runs integration tests via `tester/test.sh`, and then cleans up.
- **Build and Run**
  - **`make build`**: Builds the application Docker image.
  - **`make start`**: Builds the application and starts it using Docker Compose.
  - **`make start-sirius`**: Runs the application integrated with local Sirius. This command uses both `docker-compose.yml` and `docker-compose.sirius.yml`.
- **Clean Up**
  - **`make clean`**: Stops and removes Docker Compose resources.
- **Static Analysis**
  - **`make go-lint`**: Runs the Go linting tool inside its container.
  - **`make gosec`**: Runs the Go security scanner to check for vulnerabilities.

## Docker Compose Configuration

The project uses two Docker Compose files to manage service configurations:

- **docker-compose.yml**  
  Defines the following services:
  - **service-app**: The main application container.
  - **service-app-test**: A container running tests with a dedicated Dockerfile.
  - **localstack**: A local AWS stack emulator for S3, Secrets Manager, SQS, SSM, etc.
  - **sirius-mock**: A mock service to emulate Sirius endpoints.
  - **go-lint**: A container to run golangci-lint with preset cache and output settings.
  - **gosec**: A container to run security scans on your code.
- **docker-compose.sirius.yml**  
  Overrides for running with Sirius integration:
  - Resets some service dependencies and network settings.
  - Adjusts the `SIRIUS_BASE_URL` environment variable to point to the Sirius API.

## Usage

1. **Running Unit Tests**

   ```bash
   make test
   ```

   This command builds the test container and runs all unit tests.

2. **Running Integration Tests**

   ```bash
   make integration-test
   ```

   This will start the app, run the integration tests via the provided test script, and perform cleanup.

3. **Starting the Application**

   ```bash
   make start
   ```

   To run with Sirius integration:

   ```bash
   make start-sirius
   ```

4. **Cleaning Up**

   To stop and remove all Docker Compose resources, run:

   ```bash
   make clean
   ```

5. **For linting**

   ```bash
   make go-lint
   ```

6. **For security scanning**
   ```bash
   make gosec
   ```

## Project Setup

Before running tests or starting the application, ensure you have the required Docker images by building the project:

```bash
make build
```

The Makefile’s setup-directories target ensures that necessary directories (e.g., for test results) are created.
