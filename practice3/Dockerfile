FROM golang:1.25.6-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o myapp ./cmd/api/main.go

FROM alpine:3.20
WORKDIR /app

COPY --from=builder /app/myapp /app/myapp
COPY --from=builder /app/database/migrations /app/database/migrations

EXPOSE 8080
ENTRYPOINT ["/app/myapp"]