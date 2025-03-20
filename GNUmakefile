TESTTIMEOUT=300m
TESTARGS?=
TEST?=$$(go list ./... |grep -v 'vendor'|grep -v 'examples')

default: testacc

# Run acceptance tests
.PHONY: testacc fmt terrafmt docs tools depscheck tflint test fmtcheck lint
testacc: fmtcheck
	TF_ACC=1 go test $(TEST) -v $(TESTARGS) -timeout $(TESTTIMEOUT) -ldflags="-X=github.com/microsoft/terraform-provider-msgraph/version.ProviderVersion=acc"


fmt:
	@echo "==> Fixing source code with gofmt..."
	# This logic should match the search logic in scripts/gofmtcheck.sh
	find . -name '*.go' | grep -v vendor | xargs gofmt -s -w

terrafmt:
	@echo "==> Fixing examples with terrafmt"
	@find examples | egrep .tf | sort | while read f; do terraform fmt $$f || echo "error in $$f"; done
	@echo "==> Fixing acceptance test terraform blocks code with terrafmt..."
	@find internal | egrep "_test.go" | sort | while read f; do terrafmt fmt -f $$f; done
	@echo "==> Fixing website terraform blocks code with terrafmt..."
	@find docs | egrep .md | sort | while read f; do terrafmt fmt $$f; done
	@find templates | egrep .tmpl | sort | while read f; do terrafmt fmt $$f; done
	@find templates | egrep .md | sort | while read f; do terrafmt fmt $$f; done

docs:
	go generate

tools:
	@echo "==> installing required tooling..."
	@sh "$(CURDIR)/scripts/gogetcookie.sh"
	go install github.com/client9/misspell/cmd/misspell@latest
	go install github.com/bflad/tfproviderlint/cmd/tfproviderlint@latest
	go install github.com/bflad/tfproviderdocs@latest
	go install github.com/katbyte/terrafmt@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install mvdan.cc/gofumpt@latest
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH || $$GOPATH)/bin v1.61.0

depscheck:
	@echo "==> Checking source code with go mod tidy..."
	@go mod tidy
	@git diff --exit-code -- go.mod go.sum || \
		(echo; echo "Unexpected difference in go.mod/go.sum files. Run 'go mod tidy' command or revert any go.mod/go.sum changes and commit."; exit 1)
	@echo "==> Checking source code with go mod vendor..."
	@go mod vendor
	@git diff --compact-summary --ignore-space-at-eol --exit-code -- vendor || \
		(echo; echo "Unexpected difference in vendor/ directory. Run 'go mod vendor' command or revert any go.mod/go.sum/vendor changes and commit."; exit 1)


tflint:
	./scripts/run-tflint.sh

test: fmtcheck
	@TEST=$(TEST) ./scripts/run-test.sh

# Currently required by tf-deploy compile, duplicated by linters
fmtcheck:
	@sh "$(CURDIR)/scripts/gofmtcheck.sh"
	@sh "$(CURDIR)/scripts/timeouts.sh"
	@sh "$(CURDIR)/scripts/check-test-package.sh"

lint:
	@echo "==> Checking source code against linters..."
	@if command -v golangci-lint; then (golangci-lint run ./...); else ($(GOPATH)/bin/golangci-lint run ./...); fi