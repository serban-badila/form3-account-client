build:
	go build ./...

build-dev:
	go build -tags "unit integration load" ./...

test-unit:
	go clean -testcache && go test -race ./... -tags unit -v

test-integration:
	go clean -testcache && go test -race ./... -tags integration -v

test-load:
	go clean -testcache && go test ./... -tags load -v # the race detector slows this down and we don't want this :)

test-all: test-unit test-integration test-load
