BINARY_NAME=stash
DIST_PATH=dist

.PHONY: all clean dev base build release

all: dev

clean:
	@echo "🧹 Cleaning up..."
	rm -rf $(DIST_PATH)

dev:
	go run main.go --dry-run || exit 0

base:
	go run main.go $(filter-out $@,$(MAKECMDGOALS))

%:
	@:

build:
	goreleaser release --snapshot --clean

release:
	GITHUB_TOKEN=$$(gh auth token) goreleaser release --clean