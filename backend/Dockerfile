FROM golang:1.25-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o auth-service ./cmd/server

FROM alpine:3.20

RUN apk --no-cache add ca-certificates curl

WORKDIR /root/

COPY --from=builder /app/auth-service .
COPY --from=builder /app/keys ./keys

EXPOSE 8080
EXPOSE 9090

CMD ["./auth-service"]
