.PHONY: generate
generate:
	go generate -v ./...
	$(MAKE) goimports

.PHONY: govet
govet:
	go vet ./...

# all .go files in the ecs-init
GOFILES:=$(shell go list -f '{{$$p := .}}{{range $$f := .GoFiles}}{{$$p.Dir}}/{{$$f}} {{end}}' ./...)

.PHONY: gocyclo
gocyclo:
	# Run gocyclo over all .go files
	gocyclo -over 12 ${GOFILES}

test:
	go test -count=1 -short -v -coverprofile cover.out ./...
	go tool cover -func cover.out > coverprofile.out

GOFMTFILES:=$(shell find . -not -path './vendor/*' -type f -iregex '.*\.go')

.PHONY: importcheck
importcheck:
	$(eval DIFFS:=$(shell goimports -l $(GOFMTFILES)))
	@if [ -n "$(DIFFS)" ]; then echo "Files incorrectly formatted. Fix formatting by running goimports:"; echo "$(DIFFS)"; exit 1; fi

.PHONY: goimports
goimports:
	goimports -w $(GOFMTFILES)

.PHONY: static-check
static-check: goimports gocyclo govet importcheck
	# use default checks of staticcheck tool, except style checks (-ST*)
	# https://github.com/dominikh/go-tools/tree/master/cmd/staticcheck
	staticcheck -tests=false -checks "inherit,-ST*" ./...

