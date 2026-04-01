FROM golang:1.24-alpine AS builder

WORKDIR /app

ENV GOTOOLCHAIN=auto

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o bin/server ./cmd/server

FROM alpine:3.21

WORKDIR /app

COPY --from=builder /app/bin/server ./server

EXPOSE 8080

CMD ["./server"]
