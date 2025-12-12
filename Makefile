.PHONY:
TEST?=$$(go list ./... | grep -v 'vendor'| grep -v 'scripts'| grep -v 'version')
HOSTNAME=jameswoolfenden
FULL_PKG_NAME=github.com/jameswoolfenden/stevedore
VERSION_PLACEHOLDER=version.ProviderVersion
NAMESPACE=dev
BINARY=stevedore
OS_ARCH=darwin_amd64
TERRAFORM=./terraform/
TF_TEST=./terraform_test/

default:

build:
	go build -o ${BINARY} ./cmd/stevedore

release:
	GOOS=darwin GOARCH=amd64 go build -o ./bin/${BINARY}_${VERSION}_darwin_amd64 ./cmd/stevedore
	GOOS=freebsd GOARCH=386 go build -o ./bin/${BINARY}_${VERSION}_freebsd_386 ./cmd/stevedore
	GOOS=freebsd GOARCH=amd64 go build -o ./bin/${BINARY}_${VERSION}_freebsd_amd64 ./cmd/stevedore
	GOOS=freebsd GOARCH=arm go build -o ./bin/${BINARY}_${VERSION}_freebsd_arm ./cmd/stevedore
	GOOS=linux GOARCH=386 go build -o ./bin/${BINARY}_${VERSION}_linux_386 ./cmd/stevedore
	GOOS=linux GOARCH=amd64 go build -o ./bin/${BINARY}_${VERSION}_linux_amd64 ./cmd/stevedore
	GOOS=linux GOARCH=arm go build -o ./bin/${BINARY}_${VERSION}_linux_arm ./cmd/stevedore
	GOOS=openbsd GOARCH=386 go build -o ./bin/${BINARY}_${VERSION}_openbsd_386 ./cmd/stevedore
	GOOS=openbsd GOARCH=amd64 go build -o ./bin/${BINARY}_${VERSION}_openbsd_amd64 ./cmd/stevedore
	GOOS=solaris GOARCH=amd64 go build -o ./bin/${BINARY}_${VERSION}_solaris_amd64 ./cmd/stevedore
	GOOS=windows GOARCH=386 go build -o ./bin/${BINARY}_${VERSION}_windows_386 ./cmd/stevedore
	GOOS=windows GOARCH=amd64 go build -o ./bin/${BINARY}_${VERSION}_windows_amd64 ./cmd/stevedore

test:
	go test $(TEST) || exit 1
	echo $(TEST) | xargs -t -n4 go test $(TESTARGS) -timeout=30s -parallel=4

testacc:
	TF_ACC=1 go test $(TEST) -v $(TESTARGS) -timeout 120m


destroy:
	cd $(TERRAFORM) && terraform destroy --auto-approve


BIN=$(CURDIR)/bin
$(BIN)/%:
	@echo "Installing tools from tools/tools.go"
	@cat tools/tools.go | grep _ | awk -F '"' '{print $$2}' | GOBIN=$(BIN) xargs -tI {} go install {}

generate-docs:
	echo "does nowt"

docs:


vet:
	go vet ./...

bump:
	git push
	$(eval VERSION=$(shell git describe --tags --abbrev=0 | awk -F. '{OFS="."; $$NF+=1; print $0}'))
	git tag -a $(VERSION) -m "new release"
	git push origin $(VERSION)

psbump:
	git push
	powershell -command "./bump.ps1"

update:
	go get -u ./...
	go mod tidy
install-tools: ## Install development and security tools
	@./scripts/install-tools.sh
	pre-commit autoupdate

lint:
	golangci-lint run --fix

gci:
	gci -w .

fmt:
	gofumpt -l -w .
staticcheck: ## Run staticcheck
	@./scripts/run-staticcheck.sh

gosec: ## Run security scanner
	gosec -quiet -exclude-generated ./...

govulncheck: ## Check for known vulnerabilities
	govulncheck ./...

complexity: ## Check cyclomatic complexity
	gocyclo -over 15 -avg .

check-all: ## Run all checks (vet, staticcheck, gosec, govulncheck)
	@echo "Running all code checks..."
	@$(MAKE) vet
	@$(MAKE) staticcheck
	@$(MAKE) gosec
	@$(MAKE) govulncheck
	@echo "All checks passed!"

security-scan: ## Run security-focused checks
	@echo "Running security scans..."
	@$(MAKE) gosec
	@$(MAKE) govulncheck
	@echo "Security scan complete!"