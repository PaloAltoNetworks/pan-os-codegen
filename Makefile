all: test_go build test_sdk

build:
	go run ./cmd/codegen/main.go

test_sdk:
	cd ../generated/pango && \
	go run ./example/main.go

clean:
	rm -rf ../generated

test_go:
	go test -v ./...