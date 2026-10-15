FROM golang:1.26.6-bookworm@sha256:116d58cbd88c1297624acc6e967a060012422bacf9930927e23fb719189c6f36

RUN apt-get update && apt-get install -y netcat-openbsd libxml2-dev

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

RUN go install github.com/pact-foundation/pact-go/v2@v2.2.0 \
  && pact-go -l DEBUG install
