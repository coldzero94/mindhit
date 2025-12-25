.PHONY: generate-api build test lint run

OPENAPI_SPEC := ../../packages/protocol/tsp-output/openapi/openapi.yaml

generate-api:
	oapi-codegen -config oapi-codegen.yaml $(OPENAPI_SPEC)

build:
	go build -o ./bin/server ./cmd/server

test:
	go test -v -race -coverprofile=coverage.out ./...

lint:
	golangci-lint run

run:
	go run ./cmd/server
