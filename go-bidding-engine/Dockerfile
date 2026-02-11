# Go Bidding Engine Dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git

# Use a proxy mirror to avoid blocked Go module hosts in some networks
ENV GOPROXY=https://goproxy.io,direct
ENV GOSUMDB=off

# Copy go mod and source files
COPY go.mod ./
COPY . .

# Update go.sum and download dependencies
RUN go mod tidy && go mod download && go mod verify

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/server

# Production image
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/main .

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
CMD ["./main"]
