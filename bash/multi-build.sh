#!/bin/bash

cd ../src

go build -o ./bin/lilser
GOOS=windows GOARCH=amd64 go build -o ./bin/windows/lilser.exe
GOOS=linux GOARCH=amd64 go build -o ./bin/linux/lilser
GOOS=darwin GOARCH=arm64 go build -o ./bin/mac/lilser
