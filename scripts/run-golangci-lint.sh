#!/bin/bash
cd backend
golangci-lint run ./... \
    --enable=goimports \
    --enable=gofmt \
    --enable=vet \
    --enable=revive \
    --enable=staticcheck
