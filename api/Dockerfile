# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/server ./cmd/server

# Final runtime stage
FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/server /app/server
COPY --from=builder /app/internal/database/migrations /app/internal/database/migrations
COPY --from=builder /app/.env /app/.env

EXPOSE 8080

CMD ["/app/server"]
