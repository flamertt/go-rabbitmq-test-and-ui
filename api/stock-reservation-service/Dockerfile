# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
COPY shared/ ./shared/
COPY pkg/ ./pkg/

# Copy service files
COPY stock-reservation-service/ ./stock-reservation-service/

# Download dependencies
RUN go mod download

# Build the application
WORKDIR /app/stock-reservation-service
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates
WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/stock-reservation-service/main .

# Run the binary
CMD ["./main"] 