default: testacc

# Run acceptance tests
.PHONY: testacc
testacc:
	TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m


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