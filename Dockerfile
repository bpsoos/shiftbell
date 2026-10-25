FROM golang:1.27.3-alpine

RUN apk add --no-cache bash sqlite

WORKDIR /build

COPY go.mod go.sum ./

RUN go mod download

WORKDIR /src
