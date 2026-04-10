# Production backend image.
FROM golang:1.25.5-alpine AS builder

WORKDIR /src

COPY go.mod ./
COPY internal ./internal
COPY cmd ./cmd

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/maxapp-bot ./cmd/bot

FROM alpine:3.21

WORKDIR /app

RUN adduser -D -g '' appuser

COPY --from=builder /out/maxapp-bot /app/maxapp-bot

USER appuser

EXPOSE 3000

CMD ["/app/maxapp-bot"]
