# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /kube-ai ./cmd/kube-ai

# Final stage
FROM alpine:3.21

RUN apk add --no-cache ca-certificates curl

WORKDIR /

# Copy the binary from builder
COPY --from=builder /kube-ai /usr/local/bin/kube-ai

# Make the binary executable
RUN chmod +x /usr/local/bin/kube-ai

# Create a non-root user to run the application
RUN adduser -D -u 1000 kubeai
USER kubeai

# Run the binary
ENTRYPOINT ["kube-ai"]
CMD ["--help"] 