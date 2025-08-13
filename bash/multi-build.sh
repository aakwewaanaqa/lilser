#!/bin/zsh

cd ../src

GOOS="windows"
GOARCH="amd"
go build -o bin/windows/lilt

GOOS="linux"
GOARCH="amd"
go build -o bin/linux/lilt

GOOS="darwin"
GOARCH="arm"
go build -o bin/mac/lilt
