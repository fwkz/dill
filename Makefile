.PHONY: build fmt mkdistdir clean release $(PLATFORMS) 

PLATFORMS := darwin/amd64 linux/amd64
DIST_DIR := $(PWD)/bin
OUTPUT_BINARY := $(DIST_DIR)/dyntcp

build: mkdistdir clean fmt
	go build -o $(OUTPUT_BINARY) $(PWD)/cmd/dyntcp/main.go

fmt:
	gofmt -s -w $(PWD)

mkdistdir:
	-mkdir -p $(DIST_DIR)

clean:
	-rm $(OUTPUT_BINARY)-*

temp = $(subst /, ,$@)
os = $(word 1, $(temp))
arch = $(word 2, $(temp))

release: $(PLATFORMS)

$(PLATFORMS): mkdistdir clean fmt
	GOOS=$(os) GOARCH=$(arch) go build -o $(OUTPUT_BINARY)-$(os)-$(arch) $(PWD)/cmd/dyntcp/main.go
