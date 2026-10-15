FROM golang:1.27.1-bookworm@sha256:8d48e12ec56735e9358640898b9d9b9fcca110612ed8a5567438c0a1baa24e66

RUN apt-get update && apt-get install -y netcat-openbsd libxml2-dev

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

RUN go install github.com/pact-foundation/pact-go/v2@v2.2.0 \
  && pact-go -l DEBUG install
