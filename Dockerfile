# syntax=docker/dockerfile:1.3-labs

ARG GO
FROM golang:${GO}-bullseye

COPY <<EOF /usr/local/bin/run-test-suite
#!/bin/sh
set -e

go test -race -coverpkg=./... -coverprofile=/artifacts/coverage.out -covermode=atomic ./...
EOF

RUN chmod +x /usr/local/bin/run-test-suite

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download

COPY . ./

ENTRYPOINT [ "run-test-suite" ]