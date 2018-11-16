.PHONY: build clean run

BIN=pagerbot
GO_FLAGS=-ldflags "-extldflags '-static'"

build:
	@go get github.com/mitchellh/gox
	@gox -osarch "linux/amd64" $(GO_FLAGS) -output "dist/{{.OS}}_{{.Arch}}/$(BIN)"

clean:
	@rm -rf dist/*

run:
	@bash -ec "set -a && source .ci-runner.env && set +a && dist/linux_amd64/pagerbot -c ./config.yml"
