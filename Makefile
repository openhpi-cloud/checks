#!/usr/bin/env make -f

export GO_ENABLED=0

TARGETS = check_dhcpv6

.PHONY: test build $(TARGETS)

build: $(TARGETS)

$(TARGETS): %: ./cmd/%/main.go
	@mkdir -p ./out
	go build -v -ldflags="-s -w" -o "./out/$@" "./cmd/$@/"

test:
	@go test -race -coverprofile=coverage.out ./...
	@go tool cover -func coverage.out | tail -n 1 | awk '{ print "Total coverage: " $$3 }'

clean:
	rm -rf ./out ./dist
