BINARY_NAME=stash
DIST_PATH=dist

.PHONY: all clean dev dev-version build release

all: dev

clean:
	@echo "🧹 Cleaning up..."
	rm -rf $(DIST_PATH)
	mkdir -p $(DIST_PATH)

dev:
	go run main.go --dry-run

dev-version:
	go run main.go --version

build:
	goreleaser release --snapshot --clean

release:
	GITHUB_TOKEN=$$(gh auth token) goreleaser release --clean