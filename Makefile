-include Makefile.local

TESTARGS ?=
GENERATED_OUT_PATH=../generated

CODEGEN_LOG_LEVEL ?= error

CODEGEN_SPECS := $(shell find specs/ -name '*.yaml')
CODEGEN_SOURCES := $(shell find pkg/ cmd/ -name '*.go')
CODEGEN_TEMPLATES := $(shell find templates/ -type f)

target/codegen: $(CODEGEN_SOURCES) $(CODEGEN_TEMPLATES)
	go build -o target/codegen ./cmd/codegen

ASSETS_SRC := $(shell find assets/ -type f)
ASSETS_DST := $(patsubst assets/%,$(GENERATED_OUT_PATH)/%,$(ASSETS_SRC))

.PHONY: assets
assets: $(ASSETS_DST)

.PHONY: codegen
codegen: codegen-stamp
codegen-stamp: target/codegen $(CODEGEN_SPECS)
	CODEGEN_LOG_LEVEL=$(CODEGEN_LOG_LEVEL) ./target/codegen -config config.yaml
	touch $@

$(GENERATED_OUT_PATH)/%: assets/%
	@mkdir -p $(@D)
	cp $< $@

panos-creds.txt:
	@curl --silent -k -H "Content-Type: application/x-www-form-urlencoded" -X POST \
	    https://$(PANOS_HOSTNAME)/api/?type=keygen -d "user=$(PANOS_USERNAME)&password=$(PANOS_PASSWORD)" |xq -x //key > $@

.PHONY: test
test: test/codegen test/pango test/terraform

.PHONY: test/codegen
test/codegen:
	go test -v ./...

.PHONY: test/pango
test/pango: codegen assets
	cd $(GENERATED_OUT_PATH)/pango && \
	go test -v ./...

.PHONY: test/pango-movement
test/pango-movement: codegen assets
	cd $(GENERATED_OUT_PATH)/pango && \
	go test -v ./movement/

.PHONY: test/pango-example
test/pango-example:
	cd $(GENERATED_OUT_PATH)/pango && \
	go build example/main.go

.PHONY: test/terraform
test/terraform: test/terraform-acc test/terraform-manager

.PHONY: test/terraform-manager
test/terraform-manager: codegen assets
	cd $(GENERATED_OUT_PATH)/terraform-provider-panos/ && \
	go test -v ./internal/manager/

.PHONY: test/terraform-acc
test/terraform-acc: panos-creds.txt codegen assets
	@export PANOS_API_KEY=$(shell cat panos-creds.txt) && \
	cd $(GENERATED_OUT_PATH)/terraform/ && \
	TF_ACC=1 PANOS_HOSTNAME=$(PANOS_HOSTNAME) \
	PANOS_SKIP_VERIFY_CERTIFICATE=1 \
	PANOS_USERNAME=$(PANOS_USERNAME) PANOS_PASSWORD=$(PANOS_PASSWORD) \
	go test -v ./test $(TESTARGS) |grep -v -E "(No slog handler provided|Pango logging configured)"

.PHONY: clean
clean:
	rm -rf *-stamp target/codegen $(GENERATED_OUT_PATH)/ panos-api-key.txt

ifndef VERBOSE
.SILENT:
endif
