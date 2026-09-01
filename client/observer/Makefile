.DEFAULT_GOAL := build

.PHONY:fmt vet build build-ebpf generate

fmt:
	go fmt ./...

vet: fmt
	go vet ./...

clean:
	go clean ./...

build: vet
	go build

# Linux-only: requires clang + libbpf-dev (see bpf/gen.go) to generate the
# eBPF bindings, then builds with EBPFProcessViewer compiled in.
generate:
	go generate ./bpf/

build-ebpf: generate vet
	go build -tags ebpf