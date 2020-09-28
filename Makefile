
.PHONY: build clean help

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

test-short: ## Run tests with -short flag
	go test -timeout 30s -short -race ./...

test: ## Run tests
	go test -timeout 30s -race ./...

cover: ## Run tests and open coverage report in browser
	go test -cover -coverprofile cover.out ./...
	go tool cover -html cover.out

compile-tests: ## Compile test and benchmarks
	for pkg in $$(go list ./...) ; do \
		go test -c -bench . $$pkg ; \
	done

gofmt: ## Run gofmt locally without overwriting any file
	gofmt -d -s $$(find . -name '*.go' | grep -v vendor)

gofmt-write: ## Run gofmt locally overwriting files
	gofmt -w -s $$(find . -name '*.go' | grep -v vendor)

govet: ## Run go vet on the project
	go vet ./...

us = url_shortener
usc = url_shortener/cmd
hl = httpload
hlc = httpload/cmd
build: ## Build all
	go build -v -o build/${us}/http_server			${usc}/http_server/main.go
	go build -v -o build/${us}/end_to_end_tester	${usc}/end_to_end_tester/main.go
	go build -v -o build/${hl}/httpload				${hlc}/httpload/main.go

clean: ## Clean all
	rm -rf build

run-httpload: build ## Run httpload (example: make run-httpload ARGS="-w 50 -n 200 -z 3s http://www.google.com")
	./build/${hl}/httpload ${ARGS}
