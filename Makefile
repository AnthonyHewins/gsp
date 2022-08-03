.PHONY: clean help fmt $(targets)

.DEFAULT: help

targets := cli

$(targets):
	go build -o bin/$@ cmd/$@/*.go

fmt: ## Run gofmt
	find *.go -type f -exec gofmt -w -s {} \;

test: ## Run go vet, and test the whole repo
	go vet ./...
	go test ./...

clean: fmt # gofmt, then tidy modules, delete the bin folder, and clean go's cache
	go mod tidy
	go clean -cache -modcache -i -r
	rm -rf ./bin

help: ## Print help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
	@printf "\033[36m%-30s\033[0m %s\n" "(target)" "Build a target binary: $(targets)"
