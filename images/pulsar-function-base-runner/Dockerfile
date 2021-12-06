FROM centos:8  as base
FROM golang:1.16 as builder

WORKDIR /workspace

# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

# Copy the go source
COPY main.go main.go

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o shRuntime main.go

RUN mkdir -p /pulsar/bin/ \
    && mkdir -p /pulsar/function/ \
    && mkdir -p /pulsar/conf/ \
    mkdir -p /pulsar/logs/

COPY --from=builder  /workspace/shRuntime /pulsar/bin/shRuntime
