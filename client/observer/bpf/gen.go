// Package bpf holds the eBPF C source for the process monitor and the
// bpf2go-generated Go bindings used to load and interact with it.
//
// Regenerating requires a Linux host with clang and libbpf headers
// installed (e.g. `apt-get install clang libbpf-dev`):
//
//	go generate ./bpf/
//
// The generated bindings are only compiled into the observer binary with
// the "ebpf" build tag (see event/process_ebpf.go); without it,
// event/process_ebpf_stub.go is used instead, which lets the rest of the
// module build on any platform, including without running `go generate`.
package bpf

//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -type process_event Bpf monitor.c -- -O2 -g -Wall
