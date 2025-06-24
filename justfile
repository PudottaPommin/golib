#!/usr/bin/env just --justfile

set windows-shell := ["pwsh.exe", "-c"]

check:
    go vet ./...

fmt:
    go fmt ./...
    go tool goimports -w .

update:
    go get -u
    go mod tidy -v

test:
    go test ./... -cpu "1,6,12"

bench:
    go test ./... -bench . -benchtime 3s -benchmem -cpu "1,6,12"