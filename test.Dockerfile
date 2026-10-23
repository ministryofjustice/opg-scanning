FROM golang:1.26.10-bookworm@sha256:632dff7351799a5a9954a48dc06670ade72ee44e8c351cdc0e5a57af2fea21af

RUN apt-get update && apt-get install -y netcat-openbsd libxml2-dev

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

RUN go install github.com/pact-foundation/pact-go/v2@v2.2.0 \
  && pact-go -l DEBUG install
