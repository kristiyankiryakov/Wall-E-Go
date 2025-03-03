# Stage 1: Build the Go binary
FROM golang:1.24.0 AS builder

# Set container working dir
WORKDIR /app

# Copy the Go module files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

#Copy the project
COPY . .

# Build the binary from the cmd dir
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o main ./cmd

#Final stage
FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/main .

CMD [ "./main" ]