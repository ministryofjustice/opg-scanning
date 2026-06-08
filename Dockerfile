FROM golang:1.26-alpine@sha256:f23e8b227fb4493eabe03bede4d5a32d04092da71962f1fb79b5f7d1e6c2a17f AS build-env

RUN apk add gcc libc-dev libxml2-dev

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY internal internal
COPY main.go .

RUN CGO_ENABLED=1 go build -a -installsuffix cgo -o /go/bin/opg-scanning .

FROM alpine:3@sha256:5b10f432ef3da1b8d4c7eb6c487f2f5a8f096bc91145e68878dd4a5019afde11

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
