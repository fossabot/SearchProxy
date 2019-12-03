all: build

TESTS = test-geoip test-memcache test-mirrorsort test-server test-util/network test-util/system test-workerpool

build:
	@go mod why
	@go build -v -x -ldflags "-s -w" -o searchproxy *.go

dockerimage:
	@cp -r /usr/local/etc/openssl ./ssl
	@docker build -t tb0hdan/searchproxy .

lint:
	@golangci-lint run --enable-all --disable=gosec

test: $(TESTS)

$(TESTS):
	@go test -bench=. -v -benchmem -race ./$(shell echo $@|awk -F'test-' '{print $$2}')
