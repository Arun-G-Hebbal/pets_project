# --------------------------
# 1. Build Stage
# --------------------------
FROM golang:1.25-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy rest of the project
COPY . .

# Build binary
RUN go build -o server main.go


# --------------------------
# 2. Runtime Stage
# --------------------------
FROM alpine:3.18

WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY --from=builder /app/server .
COPY .env /app/.env

RUN mkdir -p /app/uploads

EXPOSE 8081

CMD ["./server"]
