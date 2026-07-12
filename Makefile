

BINARY_NAME := cat-run
SRC_DIR     := ./src
BIN_DIR     := ./bin

LDFLAGS := -s -w


TARGETS := 	darwin/amd64 	darwin/arm64 	linux/amd64 	linux/arm64 	windows/amd64

.PHONY: all build build-all test clean package-npm publish-dry

all: build


build:
	go build -ldflags="$(LDFLAGS)" -o $(BIN_DIR)/$(BINARY_NAME) $(SRC_DIR)


build-all: clean-bin
	@mkdir -p $(BIN_DIR)
	@for target in $(TARGETS); do 		GOOS=$$(echo $$target | cut -d/ -f1); 		GOARCH=$$(echo $$target | cut -d/ -f2); 		OUT_DIR=$(BIN_DIR)/$$GOOS-$$GOARCH; 		mkdir -p $$OUT_DIR; 		if [ "$$GOOS" = "windows" ]; then 			go build -ldflags="$(LDFLAGS)" -o $$OUT_DIR/$(BINARY_NAME).exe $(SRC_DIR); 		else 			go build -ldflags="$(LDFLAGS)" -o $$OUT_DIR/$(BINARY_NAME) $(SRC_DIR); 		fi; 		echo "✅  $$GOOS/$$GOARCH"; 	done

# Run Go tests
test:
	go test -v ./...

# Clean build artifacts
clean: clean-bin
	rm -rf npm/ dist/

clean-bin:
	rm -rf $(BIN_DIR)/

# Generate npm packages from built binaries
package-npm: build-all
	node scripts/build-npm.js

# Dry-run publish (validate without uploading)
publish-dry: package-npm
	cd npm && 	for pkg in cat-run-*/; do 		cd "$$pkg" && npm publish --dry-run && cd ..; 	done && 	cd cat-run && npm publish --dry-run