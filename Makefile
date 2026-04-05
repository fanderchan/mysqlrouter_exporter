SHELL := /bin/bash

.PHONY: test build release-artifacts clean docker-build

test:
	go test ./...

build:
	bash ./build.sh

release-artifacts:
	GOOS=linux GOARCH=amd64 bash ./build.sh
	GOOS=linux GOARCH=arm64 bash ./build.sh

clean:
	rm -rf build dist coverage.out

docker-build:
	docker build -t mysqlrouter_exporter:local .
