FROM golang:1.26-alpine@sha256:7a3e50096189ad57c9f9f865e7e4aa8585ed1585248513dc5cda498e2f41812c AS build-env

RUN apk add gcc libc-dev libxml2-dev

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY internal internal
COPY main.go .

RUN CGO_ENABLED=1 go build -a -installsuffix cgo -o /go/bin/opg-scanning .

FROM alpine:3@sha256:28bd5fe8b56d1bd048e5babf5b10710ebe0bae67db86916198a6eec434943f8b

RUN apk update && \
    apk add libxml2-dev && \
    apk upgrade --no-cache busybox libcrypto3 libssl3 musl musl-utils

ENV XSD_PATH=/go/xsd

WORKDIR /go/bin

COPY --from=build-env /go/bin/opg-scanning main
COPY xsd /go/xsd

RUN addgroup -S app && \
    adduser -S -g app app && \
    chown -R app:app main
USER app
ENTRYPOINT ["./main"]
