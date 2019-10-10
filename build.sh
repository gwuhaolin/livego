#!/usr/bin/env bash

rm -rf bin
mkdir bin

go build
mv livego ./bin