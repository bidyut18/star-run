BINARY_NAME=uni-run

.PHONY: build build-native build-all test clean package-npm

# Native build for current OS/arch
build build-native:
	go build -ldflags="-s -w" -o bin/$(BINARY_NAME) ./src

# Cross-compile for all platforms
build-all:
	mkdir -p bin/darwin-amd64 bin/darwin-arm64 bin/linux-amd64 bin/linux-arm64 bin/windows-amd64
	GOOS=darwin  GOARCH=amd64 go build -ldflags="-s -w" -o bin/darwin-amd64/$(BINARY_NAME) ./src
	GOOS=darwin  GOARCH=arm64 go build -ldflags="-s -w" -o bin/darwin-arm64/$(BINARY_NAME) ./src
	GOOS=linux   GOARCH=amd64 go build -ldflags="-s -w" -o bin/linux-amd64/$(BINARY_NAME) ./src
	GOOS=linux   GOARCH=arm64 go build -ldflags="-s -w" -o bin/linux-arm64/$(BINARY_NAME) ./src
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o bin/windows-amd64/$(BINARY_NAME).exe ./src

# Run tests
test:
	go test ./src

# Clean binaries and generated npm package.json files
clean:
	rm -rf bin/*
	rm -f npm/*/uni-run npm/*/uni-run.exe  # remove binaries in npm packages
	rm -f npm/*/package.json               # remove generated package.json files

# Package all binaries into npm tarballs
package-npm: build-all
	node scripts/build-npm.js