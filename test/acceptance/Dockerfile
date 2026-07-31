FROM golang:1.26.5-alpine AS source

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

FROM source AS application

RUN go tool templ generate && go build -o /usr/local/bin/shiftbell ./cmd/serve

ENTRYPOINT ["/usr/local/bin/shiftbell"]

FROM source AS test-runner

ENTRYPOINT ["/bin/sh", "-c", "true"]
