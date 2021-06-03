# syntax = docker/dockerfile:1.2
ARG GO_VERSION=1.16

FROM golang:${GO_VERSION} AS BUILDER

WORKDIR /effx

COPY go.mod go.mod
COPY go.sum go.sum
RUN --mount=type=ssh go mod download && go mod verify

COPY main.go main.go
COPY internal/ internal/

RUN go build -o cluster-agent ./

FROM gcr.io/distroless/base-debian10

COPY --from=BUILDER /effx/cluster-agent /usr/bin/cluster-agent

CMD [ "/usr/bin/cluster-agent" ]