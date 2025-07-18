# Build stage
FROM golang:1.21-alpine AS builder

# Install git and ca-certificates (needed for go mod download)
RUN apk add --no-cache git ca-certificates

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Copy shared dependencies
COPY shared/ ./shared/
COPY pkg/ ./pkg/

# Copy auth service source
COPY auth-service/ ./auth-service/

# Download dependencies
RUN go mod download

# Build the application
WORKDIR /app/auth-service
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS calls
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN addgroup -g 1001 appgroup && adduser -u 1001 -G appgroup -s /bin/sh -D appuser

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/auth-service/main .

# Change ownership and switch to non-root user
RUN chown appuser:appgroup main
USER appuser

# Expose port
EXPOSE 8086

# Command to run
CMD ["./main"] 