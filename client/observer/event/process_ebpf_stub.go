//go:build !ebpf

package event

import "log"

// EBPFProcessViewer stands in for the real implementation (process_ebpf.go)
// when this binary was built without the "ebpf" tag - the default
// everywhere except the Linux eBPF Docker build. Rebuild with `-tags ebpf`
// on Linux, after `go generate ./bpf/`, to get the real one.
type EBPFProcessViewer struct{}

var _ ProcessViewer = (*EBPFProcessViewer)(nil)

func NewEBPFProcessViewer(handler EventHandler) *EBPFProcessViewer {
	return &EBPFProcessViewer{}
}

func (p *EBPFProcessViewer) Start() {
	log.Fatal("eBPF support not built in: rebuild with -tags ebpf on Linux (see bpf/gen.go)")
}
