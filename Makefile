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
	@mkdir -p bin
	@$(GO_ENV) go build -o bin/nd ./cmd/nd
	@$(GO_ENV) go build -o bin/ndadm ./cmd/ndadm

fmt: ## Format all Go files
	@$(GO_ENV) go fmt ./...

lint: ## Run go vet on all packages
	@$(GO_ENV) go vet ./...

prepare: build fmt lint install-hooks ## Prepare for testing

install-hooks: ## Install git hooks
	@echo "Installing git hooks..."
	@mkdir -p .git/hooks
	@cp .githooks/pre-commit .git/hooks/pre-commit
	@chmod +x .git/hooks/pre-commit
	@git config core.hooksPath .git/hooks

test: prepare ## Run all tests
	@$(GO_ENV) go test ./features

test-discovery: prepare ## Run discovery scenarios
	@$(GO_ENV) go test ./features -run TestFeatures/Learning_about_

test-registration: prepare ## Run registration scenarios only
	@$(GO_ENV) go test ./features -run 'TestFeatures/(Successful_registration|Registration_fails_when_network_is_down|Impersonation_is_prevented|Re-registration_by_same_client_succeeds|Registration_without_name_flag|Multiple_agents_can_register)'

test-communication: prepare ## Run communication scenarios only
	@$(GO_ENV) go test ./features -run 'TestFeatures/(Sending_a_need_and_receiving_it|Intent_must_precede_solution|Successful_solution_flow)'

test-report: prepare ## Run tests and generate HTML report
	@echo "Running tests and generating report..."
	@$(GO_ENV) GODOG_FORMAT=cucumber:report.json go test ./features -v || true
	@if [ -f report.json ]; then \
		echo "Test report generated: report.json"; \
		echo "Install cucumber-html-reporter for HTML: npm install -g cucumber-html-reporter"; \
	fi

clean:
	rm -rf bin .go
