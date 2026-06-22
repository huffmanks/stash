BINARY_NAME=stash
DIST_PATH=dist
VERSION ?= 1.0.7

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

tag:
	git tag v$(VERSION)
	git push origin v$(VERSION)

release:
	GITHUB_TOKEN=$$(gh auth token) goreleaser release --clean