# Build stage
FROM golang:latest AS builder

WORKDIR /app

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o main

# Run stage
FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/main .

CMD ["/app/main"]