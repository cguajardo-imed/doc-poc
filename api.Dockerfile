# ---- Build stage ----
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -ldflags="-s -w" -o summarizer-api .

# ---- Runtime stage ----
FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/summarizer-api .
COPY --from=builder /app/files/ ./files/

EXPOSE 3000

CMD ["./summarizer-api"]
