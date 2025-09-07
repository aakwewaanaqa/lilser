#!/bin/bash

cd ../src

go build -o ./bin/lilser
GOOS=windows GOARCH=amd64 go build -o ./bin/windows-amd-lilser.exe
GOOS=linux GOARCH=amd64 go build -o ./bin/linux-amd-lilser
GOOS=darwin GOARCH=arm64 go build -o ./bin/mac-arm-lilser
