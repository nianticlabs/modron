#!/bin/bash
set -e
# Unfortunately Docker doesn't support symlinks, so we need to copy the proto files

# Make sure we're running in the script directory
pushd "$(dirname "$0")"
cp ../../proto/*.proto proto/
popd

