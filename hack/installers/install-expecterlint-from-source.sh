#!/bin/bash
set -eux -o pipefail

# This script builds expecterlint from source with the correct Go version
# to avoid compatibility issues with the released binary.

TEMP_DIR=$(mktemp -d)
trap "rm -rf $TEMP_DIR" EXIT

cd "$TEMP_DIR"
git clone https://github.com/d0ubletr0uble/expecterlint.git
cd expecterlint

# Update go.mod to use go 1.25 using go mod edit for robustness
go mod edit -go=1.25.0

# Build and install
INSTALL_PATH="${GOPATH:-$HOME/go}/bin/expecterlint"
go build -o "$INSTALL_PATH" ./cmd/expecterlint

echo "expecterlint successfully installed to $INSTALL_PATH"

# Check if the install directory is in PATH
if ! echo "$PATH" | grep -q "$(dirname "$INSTALL_PATH")"; then
    echo ""
    echo "WARNING: $(dirname "$INSTALL_PATH") is not in your PATH."
    echo "Add it to your PATH with: export PATH=\"$(dirname "$INSTALL_PATH"):\$PATH\""
fi
