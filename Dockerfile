FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go run github.com/swaggo/swag/cmd/swag@v1.8.10 init -g main.go
RUN go build -o backend main.go

# use minimal image
FROM alpine:latest

WORKDIR /app
RUN apk --no-cache add ca-certificates
COPY --from=builder /app/backend .
EXPOSE 8080
CMD ["./backend"]

