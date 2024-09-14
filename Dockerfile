# syntax=docker/dockerfile:1
FROM golang:1.23 AS builder

WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN go build -o aptforge ./cmd

# Create a minimal image
FROM debian:buster-slim

# Install required packages (if any)
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

# Copy the binary from the builder stage
COPY --from=builder /app/aptforge /usr/local/bin/aptforge

# Set the entrypoint
ENTRYPOINT ["aptforge"]
