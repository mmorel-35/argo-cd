#!/bin/bash
# apply-vendor-patches.sh
#
# Applies custom patches to the vendor directory after `go mod vendor`.
# These patches add compile-time stubs needed for grpc-gateway v2 compat
# with third-party packages that do not yet implement
# google.golang.org/protobuf/proto.Message (i.e., ProtoReflect()).
#
# Must be called after every `go mod vendor` run (see Makefile: apply-vendor-patches).

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(dirname "$SCRIPT_DIR")"
PATCHES_DIR="$SCRIPT_DIR/vendor-patches"
VENDOR_DIR="$REPO_ROOT/vendor"

echo "Applying vendor patches from $PATCHES_DIR ..."

find "$PATCHES_DIR" -type f -name "*.go" | while read -r patch_file; do
  relative="${patch_file#"$PATCHES_DIR/"}"
  dest="$VENDOR_DIR/$relative"
  dest_dir="$(dirname "$dest")"
  if [ ! -d "$dest_dir" ]; then
    echo "  WARNING: vendor package directory not found, skipping: $dest_dir"
    continue
  fi
  cp "$patch_file" "$dest"
  echo "  patched: vendor/$relative"
done

echo "Done."
