#!/bin/bash
set -e
set -u
set -o pipefail

cd $( readlink -f "$( dirname "${0}" )/.." )

# Setup temporary GOPATH so we can install go-bindata from vendor
export GOPATH=$( mktemp -d )
ln -s $( pwd )/vendor "${GOPATH}/src"
go install "./vendor/github.com/go-bindata/go-bindata/..."

OUTDIR=${OUTDIR:-"."}
output="${OUTDIR}/pkg/operator/v410_00_assets/bindata.go"
${GOPATH}/bin/go-bindata \
    -nocompress \
    -nometadata \
    -prefix "bindata" \
    -pkg "v410_00_assets" \
    -o "${output}" \
    -ignore "OWNERS" \
    bindata/v4.1.0/...
gofmt -s -w "${output}"
