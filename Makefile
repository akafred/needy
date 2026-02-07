.PHONY: $(shell sed -n -e '/^$$/ { n ; /^[^ .\#][^ ]*:/ { s/:.*$$// ; p ; } ; }' $(MAKEFILE_LIST))

# Local Go environment variables to bypass permission issues
GOCACHE := $(shell pwd)/.go/cache
GOTMPDIR := $(shell pwd)/.go/tmp
GOMODCACHE := $(shell pwd)/.go/mod
GOSUMDB := off

GO_ENV := GOCACHE=$(GOCACHE) GOTMPDIR=$(GOTMPDIR) GOMODCACHE=$(GOMODCACHE) GOSUMDB=$(GOSUMDB)

help: # Extracts make targets with doble-hash comments and prints them
	@grep -hE '^\S+:.*##' $(MAKEFILE_LIST) | sed -e 's/:.*##\s*/ : /' | while IFS=' : ' read -r cmd desc; do \
		printf "\033[36m%-20s\033[0m %s\n" "$$cmd" "$$desc"; \
	done

build: ## Build the nd and ndadm binaries
	mkdir -p bin
	$(GO_ENV) go build -o bin/nd ./cmd/nd
	$(GO_ENV) go build -o bin/ndadm ./cmd/ndadm

test: build ## Run all tests
	$(GO_ENV) go test ./features -v

test-discovery: build ## Run discovery scenarios
	$(GO_ENV) go test ./features -v -run TestFeatures/Learning_about_

clean:
	rm -rf bin .go
