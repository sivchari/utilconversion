#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

SCRIPT_ROOT=$(dirname "${BASH_SOURCE[0]}")/..

# Generate deepcopy code for both v1 and v1alpha1
echo "Generating deepcopy code..."
deepcopy-gen \
  --output-file=zz_generated.deepcopy.go \
  --go-header-file="${SCRIPT_ROOT}/hack/boilerplate.go.txt" \
  github.com/sivchari/utilconversion/_example/api/v1 \
  github.com/sivchari/utilconversion/_example/api/v1alpha1

# Generate conversion code
echo "Generating conversion code..."
conversion-gen \
  --extra-peer-dirs=github.com/sivchari/utilconversion/_example/api/v1 \
  --output-file=zz_generated.conversion.go \
  --go-header-file="${SCRIPT_ROOT}/hack/boilerplate.go.txt" \
  github.com/sivchari/utilconversion/_example/api/v1alpha1

echo "Code generation completed successfully"
