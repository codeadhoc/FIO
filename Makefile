GOCMD=go
GOBUILD=$(GOCMD) build
GORUN=$(GOCMD) run
GOINSTALL=$(GOCMD) install
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test

.PHONY: build-all # - Builds all linux arch go binary
build-all:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) ./cmd/tls/httptlsserver.go 
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) ./cmd/mtls/httpmtlsserver.go
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) ./cmd/httpinsecureserver.go 

.PHONY: build-tls # - Builds linux arch go binary for tls server
build-tls:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) ./cmd/tls/httptlsserver.go 

.PHONY: build-mtls # - Builds linux arch go binary for mtls server
build-mtls:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) ./cmd/mtls/httpmtlsserver.go

.PHONY: build-insecure # - Builds linux arch go binary insecure server
build-insecure:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) ./cmd/httpinsecureserver.go 

.PHONY: tasks
tasks:
	@grep '^.PHONY: .* #' Makefile | sed 's/\.PHONY: \(.*\) # \(.*\)/\1 \2/' | expand -t20