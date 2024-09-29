# Build stage
FROM golang:1.22-alpine AS builder

# Install build dependencies
RUN apk add --no-cache gcc musl-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go install github.com/swaggo/swag/cmd/swag@latest
RUN swag init -g cmd/rest.go

# Build the application with the correct name
# Adjust the path to your main.go file
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o moneytransfer .

# Final stage
FROM alpine:latest

# Install PostgreSQL client
RUN apk --no-cache add ca-certificates postgresql-client

WORKDIR /root/

COPY --from=builder /app/moneytransfer .
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/docs ./docs

EXPOSE 8080

# Run the moneytransfer REST service
CMD ["./moneytransfer", "rest"]