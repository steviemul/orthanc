//go:build ebpf

// This file requires the bpf2go-generated bindings in package bpf, which
// only exist after `go generate ./bpf/` has been run on a Linux host (see
// bpf/gen.go). It's built only with `-tags ebpf`; everywhere else
// process_ebpf_stub.go provides a stand-in EBPFProcessViewer.
package event

import (
	"bytes"
	"encoding/binary"
	"errors"
	"log"

	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/ringbuf"
	"github.com/cilium/ebpf/rlimit"

	"github.com/steviemul/orthanc-observer/bpf"
)

const (
	ebpfEventStarted uint32 = 0
	ebpfEventStopped uint32 = 1
)

// EBPFProcessViewer watches process start/stop via kernel tracepoints
// (sched_process_exec / sched_process_exit) instead of polling /proc.
type EBPFProcessViewer struct {
	handler EventHandler
}

var _ ProcessViewer = (*EBPFProcessViewer)(nil)

func NewEBPFProcessViewer(handler EventHandler) *EBPFProcessViewer {
	return &EBPFProcessViewer{
		handler: handler,
	}
}

func (p *EBPFProcessViewer) Start() {
	if err := rlimit.RemoveMemlock(); err != nil {
		log.Fatalf("failed to remove memlock limit: %v", err)
	}

	objs := bpf.BpfObjects{}
	if err := bpf.LoadBpfObjects(&objs, nil); err != nil {
		log.Fatalf("loading bpf objects failed: %v", err)
	}
	defer objs.Close()

	execTp, err := link.Tracepoint("sched", "sched_process_exec", objs.HandleExec, nil)
	if err != nil {
		log.Fatalf("failed to attach exec tracepoint: %v", err)
	}
	defer execTp.Close()

	exitTp, err := link.Tracepoint("sched", "sched_process_exit", objs.HandleExit, nil)
	if err != nil {
		log.Fatalf("failed to attach exit tracepoint: %v", err)
	}
	defer exitTp.Close()

	rd, err := ringbuf.NewReader(objs.Events)
	if err != nil {
		log.Fatalf("failed to open ringbuffer reader: %v", err)
	}
	defer rd.Close()

	var raw bpf.BpfProcessEvent

	for {
		record, err := rd.Read()
		if err != nil {
			if errors.Is(err, ringbuf.ErrClosed) {
				return
			}
			log.Printf("failed to read from ringbuffer: %v", err)
			continue
		}

		if err := binary.Read(bytes.NewBuffer(record.RawSample), binary.LittleEndian, &raw); err != nil {
			log.Printf("failed to parse event data: %v", err)
			continue
		}

		comm := string(bytes.TrimRight(raw.Comm[:], "\x00"))

		eventType := "STARTED"
		if raw.Type == ebpfEventStopped {
			eventType = "STOPPED"
		}

		p.handler.HandleEvent(BuildEvent(eventType, int(raw.Pid), comm))
	}
}
