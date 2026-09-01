# --- Stage 1: Build the Go Application ---
FROM golang:1.27 AS builder

WORKDIR /app

# Copy dependency files first for caching layers
COPY go.mod ./
RUN go mod download

# Copy source code
COPY . .

# Build a statically linked binary (ideal for container portability)
RUN CGO_ENABLED=0 GOOS=linux go build -o orthanc-observer .

# --- Stage 2: Final Runtime Image ---
FROM ubuntu:24.04

# Install basic debugging tools useful for process monitoring tests
RUN apt-get update && apt-get install -y \
    procps \
    curl \
    && rm -rf /var/lib/apt/lists/*

# Copy the compiled binary from the builder stage
COPY --from=builder /app/orthanc-observer /usr/local/bin/orthanc-observer

# Run the observer as the main container process
CMD ["/usr/local/bin/orthanc-observer"]