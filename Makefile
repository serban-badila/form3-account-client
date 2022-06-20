build:
	go build ./...

build-dev:
	go build -tags "unit integration" ./...

unit-test:
	go clean -testcache && go test -race ./... -tags unit -v

integration-test:
	go clean -testcache && go test -race ./... -tags integration -v

test-all: unit-test integration-test
