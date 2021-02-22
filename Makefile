GOCMD        = go
GOBUILD      = $(GOCMD) build
GOCLEAN      = $(GOCMD) clean
GOTEST       = $(GOCMD) test
GOVET        = $(GOCMD) vet
GOGET        = $(GOCMD) get
GOGET        = $(GOCMD) get

GOX_ARGS     = -output="$(BUILD_DIR)/{{.Dir}}-{{.OS}}-{{.Arch}}" -osarch="linux/arm linux/amd64 linux/arm64"

BUILD_DIR    = build
BINARY_NAME  = airquality-homekit

all: clean vet build

build:
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) -v

vet:
	${GOVET} ./...

clean:
	$(GOCLEAN)
	rm -f $(BUILD_DIR)/*

run: build
	./$(BUILD_DIR)/$(BINARY_NAME)

release:
	${GOGET} -u github.com/mitchellh/gox
	gox -ldflags "${LD_FLAGS}" ${GOX_ARGS}
	shasum -a 512 build/* > build/sha512sums.txt

.PHONY: all vet clean build run release
