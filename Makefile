.PHONY: build fmt clean

DIST_DIR=$(PWD)/bin
OUTPUT_BINARY=$(DIST_DIR)/dyntcp

build: clean fmt
	mkdir -p $(DIST_DIR)
	go build -o $(OUTPUT_BINARY) $(PWD)/cmd/dyntcp/main.go

fmt:
	gofmt -s -w $(PWD)

clean:
	-rm $(OUTPUT_BINARY)