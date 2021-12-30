#!/bin/sh

[ ! -d bin ] && mkdir bin

go build -tags netgo -o bin/req
