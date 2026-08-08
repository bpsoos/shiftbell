FROM golang:1.26.5-alpine

RUN apk add --no-cache bash

WORKDIR /build

COPY go.mod go.sum ./

RUN go mod download

WORKDIR /src
