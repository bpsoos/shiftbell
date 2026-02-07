FROM golang:1.25.7-alpine

WORKDIR /build

COPY go.mod go.sum ./

RUN go mod download

WORKDIR /src
