# Stage 1: Build the Go application
FROM golang:1.22.5-alpine3.20 AS builder

# Install necessary packages for building
RUN apk update && apk add --no-cache git

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files
COPY ./src/go.mod ./src/go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY ./src .

# Build the Go application for production
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o pgrest ./cmd/app/main.go

# Stage 2: Create the final lightweight image
FROM alpine:3.20.1

# Set the environment variables
ENV PGREST_CONFIG_PATH="/root/config/pgrest.conf"

# Set the working directory
WORKDIR /root/

# Copy the Go binary from the builder stage
COPY --from=builder /app/pgrest .

# Command to run the application
CMD ["./pgrest"]