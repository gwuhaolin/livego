#!/usr/bin/env bash

rm -rf bin
mkdir bin

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build
mv livego bin/linux-amd64-livego
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build
mv livego bin/darwin-amd64-livego
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build
mv livego.exe bin/windows-amd64-livego.exe
cp livego.cfg bin/