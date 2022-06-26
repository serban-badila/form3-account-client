FROM golang:1.18-alpine as base

    WORKDIR /app/

    COPY go.mod .
    COPY go.sum .
    RUN go mod download

    COPY account/ ./account

    ENV CGO_ENABLED="0"

    RUN go build ./...
    
FROM base as dev-test

    COPY tests/ ./tests

    RUN go build -tags "unit integration load" ./...