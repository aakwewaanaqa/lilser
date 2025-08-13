#!/bin/zsh

cd ../src

GOOS="windows"
GOARCH="amd"
go build -o bin/windows/lilser

GOOS="linux"
GOARCH="amd"
go build -o bin/linux/lilser

GOOS="darwin"
GOARCH="arm"
go build -o bin/mac/lilser
