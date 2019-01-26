## ----- VARIABLES -----
## Go module name.
MODULE = $(shell basename "$$(pwd)")
ifeq ($(shell ls -1 go.mod 2> /dev/null),go.mod)
	MODULE = $(shell cat go.mod | grep module | awk '{print $$2}')
endif

## Program version.
VERSION ?= latest
__GIT_DESC = git describe --tags
ifneq ($(shell $(__GIT_DESC) 2> /dev/null),)
	VERSION = $(shell $(__GIT_DESC) | cut -c 2-)
endif

## Custom Go linker flag.
LDFLAGS = -X $(MODULE)/internal/info.Version=$(VERSION)



## ----- TARGETS ------
## Generic:
.PHONY: default version setup install build clean run lint test review release \
        help

default: run
version: ## Show project version (derived from 'git describe').
	@echo $(VERSION)

setup: go-setup ## Set up this project on a new device.
	@echo "Configuring githooks..." && \
	 git config core.hooksPath .githooks && \
	 echo done

install: go-install ## Install project dependencies.
build: go-build ## Build project.
clean: go-clean ## Clean build artifacts.

GOENV ?= development
run: ## Run project (development).
	@GOENV="$(GOENV)" $(MAKE) go-run

lint: go-lint ## Lint and check code.
test: go-test ## Run tests.
review: go-review ## Lint code and run tests.
release: ## Release / deploy this project.
	@echo "No release procedure defined."

## Show usage for the targets in this Makefile.
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
	   awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'


## CI:
.PHONY: ci-install ci-test ci-deploy
__KB = kubectl

ci-install: dk-pull
ci-test: dk-test
	@$(__DKCMP_VER) up --no-start && $(MAKE) dk-tags
ci-deploy:
	@$(MAKE) dk-push && \
	 for deploy in $(DEPLOYS); do \
	   $(__KB) patch deployment "$$deploy" \
	     -p "{\"spec\":{\"template\":{\"metadata\":{\"annotations\":{\"date\":\"$$(date +'%s')\"}}}}}"; \
	 done


## git-secret:
.PHONY: secrets-hide secrets-reveal
secrets-hide: ## Hides modified secret files using git-secret.
	@echo "Hiding modified secret files..." && git secret hide -m

secrets-reveal: ## Reveals secret files that were hidden using git-secret.
	@echo "Revealing hidden secret files..." && git secret reveal


## Go:
.PHONY: go-deps go-bench go-setup go-install go-build go-clean go-run go-lint \
        go-test go-review

go-deps: ## Verify and tidy project dependencies.
	@echo "Verifying module dependencies..." && \
	 go mod verify && \
	 echo "Tidying module dependencies..." && \
	 go mod tidy && \
	 echo done

go-bench: ## Run benchmarks.
	@echo "Running benchmarks with 'go test -bench=.'..." && \
	 $(__TEST) -run=^$$ -bench=. -benchmem ./...

go-setup: go-install go-deps

go-install:
	@echo "Downloading module dependencies..." && \
	 go mod download && \
	 echo done

BUILDARGS = -ldflags "$(LDFLAGS)" $(BARGS)
BDIR = .
go-build:
	@echo "Building with 'go build'..." && \
	 go build $(BUILDARGS) $(BDIR) && \
	 echo done

go-clean:
	@echo "Cleaning with 'go clean'..." && \
	 go clean $(BDIR) && \
	 echo done

go-run:
	@echo "Running with 'go run'..." && \
	 go run $(BUILDARGS) $(BDIR)

go-lint:
	@if command -v goimports > /dev/null; then \
	   echo "Formatting code with 'goimports'..." && \
	   goimports -w -l . | tee /dev/stderr | xargs -0 test -z; EXIT=$$?; \
	 else \
	   echo "'goimports' not installed, skipping format step."; \
	 fi && \
	 if command -v golint > /dev/null; then \
	   echo "Linting code with 'golint'..." && \
	   golint -set_exit_status ./...; EXIT="$${EXIT:-?}"; \
	 else \
	   echo "'golint' not installed, skipping linting step."; \
	 fi && \
	 echo "Checking code with 'go vet'..." && go vet ./... && \
	 echo done && exit $$EXIT

COVERFILE = coverage.out
TTIMEOUT  = 20s
TARGS     = -race
__TEST = go test -coverprofile="$(COVERFILE)" -covermode=atomic \
                 -timeout="$(TTIMEOUT)" $(BUILDARGS) $(TARGS) \
                 ./...
go-test:
	@echo "Running tests with 'go test':" && $(__TEST)

go-review: go-lint go-test


## Docker:
.PHONY: dk-pull dk-push dk-build dk-build-push dk-clean dk-tags dk-up \
        dk-build-up dk-down dk-logs dk-test

__DK     = docker $(DKARGS)
__DKFILE = docker-compose.yml

__DKCMP  = docker-compose -f "$(__DKFILE)" $(DKARGS)
__DKCMP_VER = VERSION="$(VERSION)" $(__DKCMP)
__DKCMP_LST = VERSION=latest $(__DKCMP)
ifeq (DKENV,test)
	__DKFILE = docker-compose.test.yml
endif

dk-pull: ## Pull latest Docker images from registry.
	@echo "Pulling latest images from registry..." && \
	 $(__DKCMP_LST) pull $(SVC)

dk-push: ## Push new Docker images to registry.
	@if git describe --exact-match --tags > /dev/null 2>&1; then \
	   echo "Pushing versioned images to registry (:$(VERSION))..." && \
	   $(__DKCMP_VER) push $(SVC); \
	 fi && \
	 echo "Pushing latest images to registry (:latest)..." && \
	 $(__DKCMP_LST) push $(SVC) && \
	 echo done

dk-build: ## Build and tag Docker images.
	@echo "Building images..." && \
	 $(__DKCMP_VER) build --parallel --compress $(SVC) && \
	 echo done && $(MAKE) dk-tags

dk-clean: ## Clean up unused Docker data.
	@echo "Cleaning unused data..." && $(__DK) system prune

dk-build-push: dk-build dk-push ## Build and push new Docker images.

dk-tags: ## Tag versioned Docker images with ':latest'.
	@echo "Tagging versioned images with ':latest'..." && \
	images="$$($(__DKCMP_VER) config | egrep image | awk '{print $$2}')" && \
	for image in $$images; do \
	  if [ -z "$$($(__DK) images -q "$$image" 2> /dev/null)" ]; then \
	    continue; \
	  fi && \
	  echo "$$image" | sed -e 's/:.*$$/:latest/' | \
	    xargs $(__DK) tag "$$image"; \
	done && \
	echo done

__DK_UP = $(__DKCMP_VER) up -d
dk-up: ## Start up containerized services.
	@echo "Bringing up services..." && $(__DK_UP) $(SVC) && echo done

dk-build-up: ## Build new images, then start them.
	@echo "Building and bringing up services..." && \
	 $(__DK_UP) --build $(SVC) && \
	 echo done

dk-down: ## Shut down containerized services.
	@echo "Bringing down services..." && \
	 $(__DKCMP_VER) down $(SVC) && \
	 echo done

dk-logs: ## Show logs for containerized services.
	@$(__DKCMP_VER) logs -f $(SVC)

__DKCMP_TEST = $(__DKCMP_VER) -f docker-compose.test.yml up
dk-test: ## Test using 'docker-compose.test.yml'.
	@if [ -s docker-compose.test.yml ]; then \
	   echo "Running containerized service tests..." && \
	   $(__DKCMP_TEST); \
	 fi
