FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY common-go/go.mod common-go/go.sum /app/common-go/
COPY common-go /app/common-go

COPY connectors/downloader/go.mod connectors/downloader/go.sum /app/connectors/downloader/

WORKDIR /app/connectors/downloader

RUN go mod download

COPY connectors/downloader /app/connectors/downloader

RUN CGO_ENABLED=0 \
    GOOS=linux \
    go build \
    -o /app/bin/downloader \
    ./cmd/downloader



FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/bin/downloader /app/downloader
COPY --from=builder /app/connectors/downloader/VERSION /app/VERSION

CMD ["/app/downloader"]
