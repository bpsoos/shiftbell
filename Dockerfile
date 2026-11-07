FROM golang:1.27.3-alpine

RUN apk add --no-cache bash sqlite

ARG TARGETARCH
ARG BIOME_VERSION=2.6.6

RUN case "$TARGETARCH" in \
        amd64) biome_arch=x64; biome_sha256=1991e0812aa489da7577fa51a96ef8d52eebd3be21a958cb9fecacfeed16e51d ;; \
        arm64) biome_arch=arm64; biome_sha256=53378861c96a525c86571f9cbe66489c3561e861c6becf5ea21ad114c70a80e7 ;; \
        *) exit 1 ;; \
    esac \
    && wget -q "https://github.com/biomejs/biome/releases/download/@biomejs/biome@${BIOME_VERSION}/biome-linux-${biome_arch}-musl" -O /usr/local/bin/biome \
    && echo "$biome_sha256  /usr/local/bin/biome" | sha256sum -c - \
    && chmod +x /usr/local/bin/biome \
    && biome --version

WORKDIR /build

COPY go.mod go.sum ./

RUN go mod download

WORKDIR /src
