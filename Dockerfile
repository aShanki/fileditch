# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install required packages
RUN apk add --no-cache gcc musl-dev

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -o fileditch cmd/server/main.go

# Final stage
FROM alpine:latest

WORKDIR /app

# Install required runtime packages
RUN apk add --no-cache sqlite

# Copy the compiled binary from builder
COPY --from=builder /app/fileditch .

# Copy static files
COPY --from=builder /app/public ./public

# Create required directories
RUN mkdir -p /app/uploads /app/data

# Set executable permissions
RUN chmod +x /app/fileditch

EXPOSE 3000

CMD ["./fileditch"]