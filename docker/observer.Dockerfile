# --- Stage 1: Build the Go Application ---
FROM golang:1.27 AS builder

# clang compiles the eBPF C source (bpf/monitor.c) to bytecode; libbpf-dev
# provides <bpf/bpf_helpers.h>. No bpftool/vmlinux.h needed - the program
# only uses plain BPF helpers, no CO-RE reads of kernel structs.
RUN apt-get update && apt-get install -y --no-install-recommends \
    clang \
    libbpf-dev \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy dependency files first for caching layers
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Generate the eBPF bindings, then build with the "ebpf" tag so
# EBPFProcessViewer (backed by the generated bindings) is compiled in
# instead of the no-op stub. See bpf/gen.go and event/process_ebpf.go.
RUN go generate ./bpf/
RUN CGO_ENABLED=0 GOOS=linux go build -tags ebpf -o orthanc-observer .

# --- Stage 2: Final Runtime Image ---
FROM ubuntu:24.04

# Install basic debugging tools useful for process monitoring tests
RUN apt-get update && apt-get install -y \
    procps \
    curl \
    && rm -rf /var/lib/apt/lists/*

# Copy the compiled binary from the builder stage
COPY --from=builder /app/orthanc-observer /usr/local/bin/orthanc-observer

# Loading BPF programs and attaching tracepoints needs CAP_BPF/CAP_PERFMON
# (kernel 5.8+) at run time, granted via docker-compose.yml (cap_add), not
# here.
ENV OBSERVER_EBPF=1

# Run the observer as the main container process
CMD ["/usr/local/bin/orthanc-observer"]
