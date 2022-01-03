#!/bin/sh

[ ! -d bin ] && mkdir bin

go generate ./...
go build -tags netgo -o bin/req
