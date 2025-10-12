#!/bin/bash
set -eux -o pipefail

# renovate: datasource=go packageName=github.com/golangci/golangci-lint
GOLANGCI_LINT_VERSION=2.5.0

GO111MODULE=on go install "github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v${GOLANGCI_LINT_VERSION}"

# renovate: datasource=go packageName=github.com/d0ubletr0uble/expecterlint
EXPECTERLINT_VERSION=1.1.0

GO111MODULE=on go install "github.com/d0ubletr0uble/expecterlint/cmd/expecterlint@v${EXPECTERLINT_VERSION}"
