#!/bin/bash

# Install goimports
go install golang.org/x/tools/cmd/goimports@latest
# Install lint
go install golang.org/x/lint/golint@latest
# Install shadow
go install golang.org/x/tools/go/analysis/passes/shadow/cmd/shadow@latest
# Install golangci-lint
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.44.0
