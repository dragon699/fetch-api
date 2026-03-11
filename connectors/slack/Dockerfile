FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY common-go/go.mod common-go/go.sum /app/common-go/
COPY common-go /app/common-go

COPY connectors/slack/go.mod connectors/slack/go.sum /app/connectors/slack/

WORKDIR /app/connectors/slack

RUN go mod download

COPY connectors/slack /app/connectors/slack

RUN CGO_ENABLED=0 \
    GOOS=linux \
    go build \
    -o /app/bin/slack \
    ./cmd/slack



FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/bin/slack /app/slack
COPY --from=builder /app/connectors/slack/VERSION /app/VERSION

CMD ["/app/slack"]
