# Build stage
FROM golang:1.24-alpine AS builder

# Set working directory
WORKDIR /app

# Install build dependencies
RUN apk add --no-cache gcc musl-dev

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -a -o mail_bot .

# Final stage
FROM alpine:latest

# Add author label
LABEL authors="nyne"

# Install required dependencies for SQLite and SSL
RUN apk --no-cache add ca-certificates tzdata sqlite libc6-compat

# Set working directory
WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/mail_bot .

# Run the application
CMD ["./mail_bot"]
