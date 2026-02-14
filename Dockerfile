FROM golang:1.26-alpine AS build-env

RUN apk add gcc libc-dev libxml2-dev

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY internal internal
COPY main.go .

RUN CGO_ENABLED=1 go build -a -installsuffix cgo -o /go/bin/opg-scanning .

FROM alpine:3

RUN apk update && \
    apk add libxml2-dev && \
    apk upgrade --no-cache busybox libcrypto3 libssl3

ENV XSD_PATH=/go/xsd

WORKDIR /go/bin

COPY --from=build-env /go/bin/opg-scanning main
COPY xsd /go/xsd

RUN addgroup -S app && \
    adduser -S -g app app && \
    chown -R app:app main
USER app
ENTRYPOINT ["./main"]
