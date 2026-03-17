app := "gitmeout"
version := "dev"
ldflags := "-s -w -X main.version=" + version

build:
    go build -ldflags "{{ldflags}}" -o bin/{{app}} ./cmd/{{app}}

run: build
    ./bin/{{app}}

test:
    go test ./...

test-verbose:
    go test -v ./...

test-coverage:
    go test -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out -o coverage.html
    echo "Coverage report: coverage.html"

lint:
    golangci-lint run

fmt:
    gofmt -w .

vet:
    go vet ./...

tidy:
    go mod tidy

validate: fmt vet lint test
    echo "All checks passed!"

clean:
    rm -rf bin/
    rm -f coverage.out coverage.html

docker-build:
    docker build -t {{app}}:{{version}} .

docker-run:
    docker run --rm -v $(pwd)/config.yaml:/app/config.yaml:ro {{app}}:{{version}}

docker-push *VERSION:
    #!/usr/bin/env bash
    version="{{VERSION}}"
    if [ -z "$version" ]; then
        version="{{version}}"
    fi
    docker build --build-arg VERSION=$version -t code.decastro.me/guilherme/gitmeout:$version -t code.decastro.me/guilherme/gitmeout:latest .
    docker push code.decastro.me/guilherme/gitmeout:$version
    docker push code.decastro.me/guilherme/gitmeout:latest

docker-push-ghcr *VERSION:
    #!/usr/bin/env bash
    version="{{VERSION}}"
    if [ -z "$version" ]; then
        version="{{version}}"
    fi
    docker build --build-arg VERSION=$version -t ghcr.io/guilycst/gitmeout:$version -t ghcr.io/guilycst/gitmeout:latest .
    docker push ghcr.io/guilycst/gitmeout:$version
    docker push ghcr.io/guilycst/gitmeout:latest

release *VERSION:
    #!/usr/bin/env bash
    version="{{VERSION}}"
    if [ -z "$version" ]; then
        echo "Usage: just release <version>"
        exit 1
    fi
    git tag -a v$version -m "Release v$version"
    git push origin v$version

install: build
    cp bin/{{app}} /usr/local/bin/{{app}}

.PHONY: build run test test-verbose test-coverage lint fmt vet tidy validate clean docker-build docker-run docker-push release install
