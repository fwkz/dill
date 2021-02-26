.PHONY: build fmt mkdistdir clean image release $(PLATFORMS) 
VERSION := $(shell git describe --tags)
PLATFORMS := darwin/amd64 linux/amd64
DIST_DIR := $(PWD)/bin
OUTPUT_BINARY := $(DIST_DIR)/dill-$(VERSION)

build: mkdistdir clean fmt
	go build -o $(OUTPUT_BINARY) $(PWD)/cmd/dill/main.go

fmt:
	gofmt -s -w $(PWD)

mkdistdir:
	-mkdir -p $(DIST_DIR)

clean:
	-rm $(OUTPUT_BINARY)-*

image:
	docker build -t dill:$(VERSION) .
	docker tag dill:$(VERSION) dill:latest 

temp = $(subst /, ,$@)
os = $(word 1, $(temp))
arch = $(word 2, $(temp))

release: $(PLATFORMS)

$(PLATFORMS): mkdistdir clean fmt
	GOOS=$(os) GOARCH=$(arch) go build -o $(OUTPUT_BINARY)-$(os)-$(arch) $(PWD)/cmd/dill/main.go

.PHONY: dill
dill: build
	./bin/dill -config config.toml
