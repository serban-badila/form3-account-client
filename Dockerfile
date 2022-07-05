FROM golang:1.18-alpine as base

    WORKDIR /app/

    COPY go.mod .
    COPY go.sum .
    RUN go mod download

    COPY account/ ./account

    # no C compiler present so the race detector and make are also missing 
    ENV CGO_ENABLED="0"

    RUN go build ./...
    
FROM base as dev-test

    COPY tests/ ./tests

    RUN go build -tags "unit integration load" ./...