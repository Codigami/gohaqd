#!/bin/bash

set -e
set -x

PACKAGE_LIST=$(go list ./... | grep -v /vendor/)

go get -u -v github.com/golang/lint/golint

go fmt $PACKAGE_LIST
go vet -v $PACKAGE_LIST
echo "$PACKAGE_LIST" | xargs -L1 golint -set_exit_status
go test $PACKAGE_LIST

