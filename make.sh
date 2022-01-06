#!/bin/sh

[ ! -d bin ] && mkdir bin

go generate ./...
go test -cover ./...
go build -tags netgo -o bin/req
