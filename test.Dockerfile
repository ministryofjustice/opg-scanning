FROM golang:1.26-bookworm@sha256:dc8e692a67c88c1a76afc36bea0814e8be5ab0d37210695f19e47cf136e0fcd5

RUN apt-get update && apt-get install -y netcat-openbsd libxml2-dev

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

RUN go install github.com/pact-foundation/pact-go/v2@v2.2.0 \
  && pact-go -l DEBUG install
