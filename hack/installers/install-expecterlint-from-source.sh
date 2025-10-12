#!/bin/bash
set -eux -o pipefail

# This script builds expecterlint from source with the correct Go version
# to avoid compatibility issues with the released binary.

TEMP_DIR=$(mktemp -d)
trap "rm -rf $TEMP_DIR" EXIT

cd "$TEMP_DIR"
git clone https://github.com/d0ubletr0uble/expecterlint.git
cd expecterlint

# Update go.mod to use go 1.25
sed -i 's/go 1.24.0/go 1.25.0/' go.mod

# Build and install
go build -o "${GOPATH:-$HOME/go}/bin/expecterlint" ./cmd/expecterlint

echo "expecterlint successfully installed to ${GOPATH:-$HOME/go}/bin/expecterlint"
