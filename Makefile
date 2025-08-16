FLAGS=
ifeq ($OS, Windows NT)
	FLAGS += -ldflags="-H windowsgui"
endif
.PHONY: all build clean install package test

all: clean build test

build:
	@echo "Building..."
	@go generate -v ./...
	@go build -v ${FLAGS} -o ./bin/ ./cmd/...

clean:
	@rm -rf ./bin
	@echo "Build artifacts cleaned"

install:
	@go install -v ./cmd/...
	@echo "Typekit CLI installed"

test:
	@go test -v ./...

package:
	@tar -czvf ./dist/typekit-cli.tar.gz ./bin/*


