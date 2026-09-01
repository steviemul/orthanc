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
	"fmt"
	"log"
	"syscall"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/ringbuf"
	"github.com/cilium/ebpf/rlimit"

	"github.com/steviemul/orthanc-observer/bpf"
)

const (
	ebpfEventStarted uint32 = 0
	ebpfEventStopped uint32 = 1
)

// int8sToBytes converts the generated bindings' [16]int8 comm field (C's
// char is signed on the target platform) into a []byte for string handling.
func int8sToBytes(s []int8) []byte {
	b := make([]byte, len(s))
	for i, v := range s {
		b[i] = byte(v)
	}
	return b
}

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

	spec, err := bpf.LoadBpf()
	if err != nil {
		log.Fatalf("loading bpf spec failed: %v", err)
	}

	if err := configureTargetPidNamespace(spec); err != nil {
		log.Fatalf("failed to configure target pid namespace: %v", err)
	}

	objs := bpf.BpfObjects{}
	if err := spec.LoadAndAssign(&objs, nil); err != nil {
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

		comm := string(bytes.TrimRight(int8sToBytes(raw.Comm[:]), "\x00"))

		eventType := "STARTED"
		if raw.Type == ebpfEventStopped {
			eventType = "STOPPED"
		}

		p.handler.HandleEvent(BuildEvent("EBPF", eventType, int(raw.Pid), comm))
	}
}

// configureTargetPidNamespace tells monitor.c which PID namespace to
// resolve traced processes' pids into: our own, identified by the dev/inode
// of /proc/self/ns/pid. Without this, the pid on each event would be the
// one bpf_get_current_pid_tgid() returns by default - the pid as seen from
// the outermost (root) PID namespace of the machine - which won't match
// anything under our own /proc, so plugins that shell out to
// /proc/<pid>/cmdline (see plugin.readCmdLineArgs) find nothing.
//
// This must be set on the CollectionSpec before it's loaded: Variable.Set
// on an already-loaded object has no effect on the running program.
func configureTargetPidNamespace(spec *ebpf.CollectionSpec) error {
	var stat syscall.Stat_t

	if err := syscall.Stat("/proc/self/ns/pid", &stat); err != nil {
		return fmt.Errorf("stat /proc/self/ns/pid: %w", err)
	}

	if err := setBpfVariable(spec, "target_ns_dev", uint64(stat.Dev)); err != nil {
		return err
	}

	return setBpfVariable(spec, "target_ns_ino", uint64(stat.Ino))
}

func setBpfVariable(spec *ebpf.CollectionSpec, name string, value uint64) error {
	v, ok := spec.Variables[name]

	if !ok {
		return fmt.Errorf("bpf variable %q not found in spec", name)
	}

	return v.Set(value)
}
