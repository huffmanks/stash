binary_name := "stash"
dist_path := "dist"
version := "1.1.0"

default: dev

clean:
    rm -rf {{dist_path}}

dev *args:
    go run main.go {{args}}

build:
    goreleaser release --snapshot --clean

tag ver=version:
    git tag v{{ver}}
    git push origin v{{ver}}

release:
    GITHUB_TOKEN=$(gh auth token) goreleaser release --clean
