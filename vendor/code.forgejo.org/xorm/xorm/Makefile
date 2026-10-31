IMPORT := code.forgejo.org/xorm/xorm
export GO111MODULE=on

GO ?= go
GOFMT ?= gofmt -s
TAGS ?=
SED_INPLACE := sed -i

GO_DIRS := caches contexts core dialects internal log migrate names schemas tags
GOFILES := $(wildcard *.go)
GOFILES += $(shell find $(GO_DIRS) -name "*.go" -type f)
INTEGRATION_PACKAGES := code.forgejo.org/xorm/xorm/tests
PACKAGES ?= $(filter-out $(INTEGRATION_PACKAGES),$(shell $(GO) list ./...))

TEST_MYSQL_HOST ?= mysql:3306
TEST_MYSQL_CHARSET ?= utf8
TEST_MYSQL_DBNAME ?= xorm_test
TEST_MYSQL_USERNAME ?= root
TEST_MYSQL_PASSWORD ?=

TEST_PGSQL_HOST ?= pgsql:5432
TEST_PGSQL_SCHEMA ?=
TEST_PGSQL_DBNAME ?= xorm_test
TEST_PGSQL_USERNAME ?= postgres
TEST_PGSQL_PASSWORD ?= mysecretpassword

TEST_CACHE_ENABLE ?= false
TEST_QUOTE_POLICY ?= always

ROOT_DIR := $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))
TEST_COVERAGE_PKG ?= code.forgejo.org/xorm/xorm/...
TEST_COVERAGE_DIR ?= $(ROOT_DIR)/coverage.out

GOFUMPT_PACKAGE ?= mvdan.cc/gofumpt@v0.12.0 # renovate: datasource=go
GOLANGCI_LINT_PACKAGE ?= github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 # renovate: datasource=go

.PHONY: all
all: build

.PHONY: build
build: go-check $(GO_SOURCES)
	$(GO) build $(PACKAGES)

.PHONY: clean
clean:
	$(GO) clean -i ./...
	rm -rf *.sql *.log test.db cover.out cover.html *coverage.out coverage.all integrations/*.sql

.PHONY: coverage-report
coverage-report:
	$(eval COVERAGE_DIRS := $(shell find $(TEST_COVERAGE_DIR) -type d | tr '\n' ',' | sed 's/,$$//'))
	$(GO) tool covdata percent -i=$(COVERAGE_DIRS)
	$(GO) tool covdata textfmt -i=$(COVERAGE_DIRS) -o=$(ROOT_DIR)/merged.coverage.out
	$(GO) tool cover -html=$(ROOT_DIR)/merged.coverage.out -o $(ROOT_DIR)/cover.html

.PHONY: fmt
fmt:
	$(GO) run $(GOFUMPT_PACKAGE) -w $(GOFILES)

.PHONY: fmt-check
fmt-check: fmt
	@git diff --exit-code --color=always $(GOFILES) \
	|| (code=$$?; echo "Please run 'make fmt' and commit the result"; exit $${code})

.PHONY: go-check
go-check:
	$(eval GO_VERSION := $(shell printf "%03d%03d%03d" $(shell go version | grep -Eo '[0-9]+\.?[0-9]+?\.?[0-9]?\s' | tr '.' ' ');))
	@if [ "$(GO_VERSION)" -lt "001011000" ]; then \
		echo "Gitea requires Go 1.11.0 or greater to build. You can get it at https://golang.org/dl/"; \
		exit 1; \
	fi
	mkdir -p $(TEST_COVERAGE_DIR)

.PHONY: help
help:
	@echo "Make Routines:"
	@echo " -                   equivalent to \"build\""
	@echo " - build             creates the entire project"
	@echo " - clean             delete integration files and build files but not css and js files"
	@echo " - fmt               format the code"
	@echo " - lint            	run code linter"
	@echo " - lint-fix          run code linter and fix issues"
	@echo " - test       		run default unit test"
	@echo " - test-mysql        run integration tests for mysql"
	@echo " - test-postgres     run integration tests for postgres"
	@echo " - test-sqlite3      run integration tests for sqlite"
	@echo " - vet               examines Go source code and reports suspicious constructs"

.PHONY: lint
lint:
	$(GO) run $(GOLANGCI_LINT_PACKAGE) run \
	|| (code=$$?; echo "Please run 'make lint-fix' and commit the result"; exit $${code})

.PHONY: lint-fix
lint-fix:
	$(GO) run $(GOLANGCI_LINT_PACKAGE) run --fix

.PHONY: test
test: go-check
	GOTOOLCHAIN=$(shell go version | awk '{print $$3}')+auto $(GO) test -cover -covermode=atomic -coverpkg=$(TEST_COVERAGE_PKG) $(PACKAGES) -args -test.gocoverdir=$(TEST_COVERAGE_DIR)

.PNONY: test-mysql
test-mysql: go-check
	$(GO) test $(INTEGRATION_PACKAGES) -v -race -db=mysql -cache=$(TEST_CACHE_ENABLE) -quote=$(TEST_QUOTE_POLICY) \
	-conn_str="$(TEST_MYSQL_USERNAME):$(TEST_MYSQL_PASSWORD)@tcp($(TEST_MYSQL_HOST))/$(TEST_MYSQL_DBNAME)?charset=$(TEST_MYSQL_CHARSET)" \
	-cover -covermode=atomic -coverpkg=$(TEST_COVERAGE_PKG) -timeout=20m -args -test.gocoverdir=$(TEST_COVERAGE_DIR)

.PHONY: test-mysql\#%
test-mysql\#%: go-check
	$(GO) test $(INTEGRATION_PACKAGES) -v -race -run $* -db=mysql -cache=$(TEST_CACHE_ENABLE) -quote=$(TEST_QUOTE_POLICY) \
	-conn_str="$(TEST_MYSQL_USERNAME):$(TEST_MYSQL_PASSWORD)@tcp($(TEST_MYSQL_HOST))/$(TEST_MYSQL_DBNAME)?charset=$(TEST_MYSQL_CHARSET)" \
	-cover -covermode=atomic -coverpkg=$(TEST_COVERAGE_PKG) -args -test.gocoverdir=$(TEST_COVERAGE_DIR)

.PNONY: test-mysql-tls
test-mysql-tls: go-check
	$(GO) test $(INTEGRATION_PACKAGES) -v -race -db=mysql -cache=$(TEST_CACHE_ENABLE) -quote=$(TEST_QUOTE_POLICY) \
	-conn_str="$(TEST_MYSQL_USERNAME):$(TEST_MYSQL_PASSWORD)@tcp($(TEST_MYSQL_HOST))/$(TEST_MYSQL_DBNAME)?charset=$(TEST_MYSQL_CHARSET)&tls=skip-verify" \
	-cover -covermode=atomic -coverpkg=$(TEST_COVERAGE_PKG) -timeout=20m -args -test.gocoverdir=$(TEST_COVERAGE_DIR)

.PHONY: test-mysql-tls\#%
test-mysql-tls\#%: go-check
	$(GO) test $(INTEGRATION_PACKAGES) -v -race -run $* -db=mysql -cache=$(TEST_CACHE_ENABLE) -quote=$(TEST_QUOTE_POLICY) \
	-conn_str="$(TEST_MYSQL_USERNAME):$(TEST_MYSQL_PASSWORD)@tcp($(TEST_MYSQL_HOST))/$(TEST_MYSQL_DBNAME)?charset=$(TEST_MYSQL_CHARSET)&tls=skip-verify" \
	-cover -covermode=atomic -coverpkg=$(TEST_COVERAGE_PKG) -args -test.gocoverdir=$(TEST_COVERAGE_DIR)

.PNONY: test-postgres
test-postgres: go-check
	$(GO) test $(INTEGRATION_PACKAGES) -v -race -db=postgres -schema='$(TEST_PGSQL_SCHEMA)' -cache=$(TEST_CACHE_ENABLE) \
	-conn_str="postgres://$(TEST_PGSQL_USERNAME):$(TEST_PGSQL_PASSWORD)@$(TEST_PGSQL_HOST)/$(TEST_PGSQL_DBNAME)?sslmode=disable" \
	-quote=$(TEST_QUOTE_POLICY) -cover -covermode=atomic -coverpkg=$(TEST_COVERAGE_PKG) -timeout=20m -args -test.gocoverdir=$(TEST_COVERAGE_DIR)

.PHONY: test-postgres\#%
test-postgres\#%: go-check
	$(GO) test $(INTEGRATION_PACKAGES) -v -race -run $* -db=postgres -schema='$(TEST_PGSQL_SCHEMA)' -cache=$(TEST_CACHE_ENABLE) \
	-conn_str="postgres://$(TEST_PGSQL_USERNAME):$(TEST_PGSQL_PASSWORD)@$(TEST_PGSQL_HOST)/$(TEST_PGSQL_DBNAME)?sslmode=disable" \
	-quote=$(TEST_QUOTE_POLICY) -cover -covermode=atomic -coverpkg=$(TEST_COVERAGE_PKG) -args -test.gocoverdir=$(TEST_COVERAGE_DIR)

.PHONY: test-sqlite3
test-sqlite3: go-check
	$(GO) test $(INTEGRATION_PACKAGES) -v -race -cache=$(TEST_CACHE_ENABLE) -db=sqlite3 -conn_str="./test.db?cache=shared&mode=rwc" \
	 -quote=$(TEST_QUOTE_POLICY) -cover -covermode=atomic -coverpkg=$(TEST_COVERAGE_PKG) -timeout=20m -args -test.gocoverdir=$(TEST_COVERAGE_DIR)

.PHONY: test-sqlite3-schema
test-sqlite3-schema: go-check
	$(GO) test $(INTEGRATION_PACKAGES) -v -race -schema=xorm -cache=$(TEST_CACHE_ENABLE) -db=sqlite3 -conn_str="./test.db?cache=shared&mode=rwc" \
	 -quote=$(TEST_QUOTE_POLICY) -cover -covermode=atomic -coverpkg=$(TEST_COVERAGE_PKG) -timeout=20m -args -test.gocoverdir=$(TEST_COVERAGE_DIR)

.PHONY: test-sqlite3\#%
test-sqlite3\#%: go-check
	$(GO) test $(INTEGRATION_PACKAGES) -v -race -run $* -cache=$(TEST_CACHE_ENABLE) -db=sqlite3 -conn_str="./test.db?cache=shared&mode=rwc" \
	 -quote=$(TEST_QUOTE_POLICY) -cover -covermode=atomic -coverpkg=$(TEST_COVERAGE_PKG) -timeout=20m -args -test.gocoverdir=$(TEST_COVERAGE_DIR)

.PNONY: test-pgx
test-pgx: go-check
	$(GO) test $(INTEGRATION_PACKAGES) -v -race -db=pgx -schema='$(TEST_PGSQL_SCHEMA)' -cache=$(TEST_CACHE_ENABLE) \
	-conn_str="postgres://$(TEST_PGSQL_USERNAME):$(TEST_PGSQL_PASSWORD)@$(TEST_PGSQL_HOST)/$(TEST_PGSQL_DBNAME)?sslmode=disable" \
	-quote=$(TEST_QUOTE_POLICY) -cover -covermode=atomic -coverpkg=$(TEST_COVERAGE_PKG) -timeout=20m -args -test.gocoverdir=$(TEST_COVERAGE_DIR)

.PHONY: test-pgx\#%
test-pgx\#%: go-check
	$(GO) test $(INTEGRATION_PACKAGES) -v -race -run $* -db=pgx -schema='$(TEST_PGSQL_SCHEMA)' -cache=$(TEST_CACHE_ENABLE) \
	-conn_str="postgres://$(TEST_PGSQL_USERNAME):$(TEST_PGSQL_PASSWORD)@$(TEST_PGSQL_HOST)/$(TEST_PGSQL_DBNAME)?sslmode=disable" \
	-quote=$(TEST_QUOTE_POLICY) -cover -covermode=atomic -coverpkg=$(TEST_COVERAGE_PKG) -timeout=20m -args -test.gocoverdir=$(TEST_COVERAGE_DIR)
