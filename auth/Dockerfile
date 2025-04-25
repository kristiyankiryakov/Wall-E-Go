# Build stage
FROM golang:1.24-alpine AS builder

# Set working directory
WORKDIR /app

# Install dependencies for build
RUN apk add --no-cache git

# Copy go.mod and go.sum files
COPY go.mod go.sum* ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-s -w" -o auth-service .

# Final stage
FROM alpine:latest

# Add necessary certificates and timezone data
RUN apk --no-cache add ca-certificates tzdata && \
    mkdir /app

WORKDIR /app

# Copy the binary from the build stage
COPY --from=builder /app/auth-service /app/

# Create a non-root user to run the application
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
RUN chown -R appuser:appgroup /app
USER appuser

# Expose the service port
EXPOSE 50051

# Command to run the service
CMD ["/app/auth-service", "serve"]