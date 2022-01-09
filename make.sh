#!/bin/sh

_version() {
	git log --decorate=full --format=format:%d |
		head -1 |
		tr ',' '\n' |
		grep tag: |
		cut -d / -f 3 |
		tr -d ',)'
}

module="$(head -1 go.mod | awk '{ print $2 }')"
version="$(_version)"

[ "$version" = "" ] && {
	version="devel $(git log -n 1 --format='format: +%h %cd' HEAD)"
}

[ ! -d bin ] && mkdir bin

go generate ./...
go test -cover ./...
go build -ldflags "-X '${module}/version.Build=$version'" -tags netgo -o bin/req
