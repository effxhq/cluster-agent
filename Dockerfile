# syntax = docker/dockerfile:1.2
ARG GO_VERSION=1.16

FROM golang:${GO_VERSION}

WORKDIR /effx

COPY go.mod go.mod
COPY go.sum go.sum
COPY vendor/ vendor/
# RUN --mount=type=ssh go mod download && go mod verify

COPY main.go main.go
COPY internal/ internal/

RUN go build -o cluster-agent ./