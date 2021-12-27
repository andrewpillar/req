#!/bin/sh

[ ! -d bin ] && mkdir bin

go build -o bin/req
