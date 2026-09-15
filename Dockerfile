ARG COREVERSION="latest"

FROM golang:1.26-bookworm AS build
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . ./

# to get buildinfo in golang
RUN apt-get update && apt-get -y install git && rm -rf /var/lib/apt/lists/*
ARG VERSION="dev"
RUN go build -v -ldflags "-X main.version=${VERSION}" ./cmd/fireantelope

####

FROM ghcr.io/streamingfast/firehose-core:${COREVERSION} AS core

####

FROM ubuntu:24.04

ARG TARGETARCH

ENV PATH="$PATH:/app"

#COPY tools/fireeth/motd_generic /etc/motd
#COPY tools/fireeth/99-fireeth.sh /etc/profile.d/
#RUN echo ". /etc/profile.d/99-fireeth.sh" > /root/.bash_aliases

RUN apt-get update && apt-get -y upgrade && apt-get -y install \
        ca-certificates htop iotop sysstat \
        strace lsof curl jq tzdata bash \
    && rm -rf /var/lib/apt/lists/*

RUN mkdir -p /app/ && curl -Lo /app/grpc_health_probe https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/v0.4.12/grpc_health_probe-linux-${TARGETARCH} && chmod +x /app/grpc_health_probe

WORKDIR /app

COPY --from=build /app/fireantelope /app/fireantelope
COPY --from=core /app/firecore /app/firecore

ENTRYPOINT ["/app/fireantelope"]
