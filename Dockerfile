FROM golang:1.26.5-alpine

WORKDIR /build

COPY go.mod go.sum ./

RUN go mod download

WORKDIR /src
