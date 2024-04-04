all: build test_sdk

build:
	go run ./cmd/codegen/main.go

test_sdk:
	cd ../generated/pango && \
	go run ./example/main.go

clean:
	rm -rf ../generated