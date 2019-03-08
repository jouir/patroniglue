#!/bin/bash
BINARY=patroniglue
VERSION=$(cat VERSION)
BUILD_PATH=/tmp/${BINARY}-${VERSION}
ldflags="-X main.AppVersion=${VERSION}"
GOOS=linux
GOARCH=amd64
DEPENDENCIES="github.com/gorilla/mux gopkg.in/yaml.v2"

export GOOS
export GOARCH

go get ${DEPENDENCIES}

go build -ldflags "$ldflags" -o ${BUILD_PATH}/${BINARY} src/*.go
(cd ${BUILD_PATH} && tar czf ${BINARY}-${VERSION}-${GOOS}-${GOARCH}.tar.gz ${BINARY})

echo "Archive created:"
ls -l ${BUILD_PATH}/${BINARY}-${VERSION}-${GOOS}-${GOARCH}.tar.gz
