BINARY_NAME=stash
DIST_PATH=dist

.PHONY: all build clean dev

all: clean build

build:
	@echo "🔨 Building binaries..."

	# MacOS
	GOOS=darwin GOARCH=amd64 go build -o $(DIST_PATH)/$(BINARY_NAME)-darwin-amd64 main.go
	GOOS=darwin GOARCH=arm64 go build -o $(DIST_PATH)/$(BINARY_NAME)-darwin-arm64 main.go

	# Linux
	GOOS=linux GOARCH=amd64 go build -o $(DIST_PATH)/$(BINARY_NAME)-linux-amd64 main.go
	GOOS=linux GOARCH=arm64 go build -o $(DIST_PATH)/$(BINARY_NAME)-linux-arm64 main.go

	@echo "✅ Done! Binaries are in the $(DIST_PATH) folder."

clean:
	@echo "🧹 Cleaning up..."
	rm -rf $(DIST_PATH)
	mkdir -p $(DIST_PATH)

dev:
	go run main.go --dry-run