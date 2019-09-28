#!/bin/bash

export GOARCH=amd64
for OS in linux darwin; do
    export GOOS=$OS
    file=dist/terraform-provider-mikrotik-${TRAVIS_TAG}-${GOOS}_${GOARCH}
    go build -o ${file}
    sha256sum ${file} > "${file}.sha256sum"
done
