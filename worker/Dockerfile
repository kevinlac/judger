FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/worker ./cmd/server
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/syncer ./cmd/syncer

FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/worker /app/worker
COPY --from=builder /app/syncer /app/syncer
ENTRYPOINT []
CMD ["/app/worker"]