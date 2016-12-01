TEST?=$$(go list ./... | grep -v vendor)
VETARGS?=-all
GOFMT_FILES?=$$(find . -name '*.go' | grep -v vendor)

default: test vet

test: vet
	govendor test $(TEST) $(TESTARGS) -v -timeout=30m -parallel=4 ;

vet: fmt
	@echo "go tool vet $(VETARGS) ."
	@go tool vet $(VETARGS) $$(ls -d */ | grep -v vendor) ; if [ $$? -eq 1 ]; then \
		echo ""; \
		echo "Vet found suspicious constructs. Please check the reported constructs"; \
		echo "and fix them if necessary before submitting the code for review."; \
		exit 1; \
	fi

fmt:
	gofmt -s -l -w $(GOFMT_FILES)

docker-test:
	sh -c "'$(CURDIR)/scripts/build_on_docker.sh' 'test'"

.PHONY: default test vet fmt
