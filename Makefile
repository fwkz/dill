.PHONY: build fmt mkdistdir clean image release $(PLATFORMS) 

VERSION := $(shell git describe --tags)
PLATFORMS := \
	darwin/amd64 \
	darwin/arm64 \
	linux/386 \
	linux/amd64 \
	linux/arm \
	linux/arm64 \
	windows/386/.exe \
	windows/amd64/.exe
DIST_DIR := $(PWD)/dist
OUTPUT_BINARY := $(DIST_DIR)/dill_$(VERSION)

build: mkdistdir clean fmt
	go build -o $(OUTPUT_BINARY) $(PWD)/cmd/dill/main.go

fmt:
	gofmt -s -w $(PWD)

mkdistdir:
	-mkdir -p $(DIST_DIR)

clean:
	-rm $(DIST_DIR)/*

image:
	docker build -t dill:$(VERSION) .
	docker tag dill:$(VERSION) dill:latest 

temp = $(subst /, ,$@)
os = $(word 1, $(temp))
arch = $(word 2, $(temp))
ext = $(word 3, $(temp))

release: $(PLATFORMS) image
	$(shell cd $(DIST_DIR); shasum -a 256 * > dill_$(VERSION)_sha256_checksums.txt)

$(PLATFORMS): mkdistdir clean fmt
	GOOS=$(os) GOARCH=$(arch) go build -o $(OUTPUT_BINARY)_$(os)_$(arch)$(ext) $(PWD)/cmd/dill/main.go

.PHONY: dill
dill: build
	$(shell $(OUTPUT_BINARY) -config config.toml)
