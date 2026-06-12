FROM golang:1.26-bookworm@sha256:13e7249b4618c115a175ea2627213131855233ecf465328cac30a0f754beb985

RUN apt-get update && apt-get install -y netcat-openbsd libxml2-dev

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

RUN go install github.com/pact-foundation/pact-go/v2@v2.2.0 \
  && pact-go -l DEBUG install
