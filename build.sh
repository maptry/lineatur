#!/bin/sh

GOOS=windows GOARCH=amd64 go build -o lineatur-win-amd64
GOOS=darwin GOARCH=amd64 go build -o lineatur-mac-amd64
GOOS=darwin GOARCH=arm64 go build -o lineatur-mac-arm64
GOOS=linux GOARCH=amd64 go build -o lineatur-linux-amd64
GOOS=linux GOARCH=arm64 go build -o lineatur-linux-arm64
