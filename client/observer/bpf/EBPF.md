# eBPF process monitor

This document explains the eBPF-based `ProcessViewer` implementation in detail: the kernel-side program in [monitor.c](monitor.c) and the userspace Go code in [../event/process_ebpf.go](../event/process_ebpf.go) that loads it and consumes its events. It's written as a companion to those two files — expect to read them side by side with this document.

If you're new to eBPF, the short version is: you write a small, restricted C program, the kernel verifies it can't crash or hang, then it runs *inside the kernel* at a chosen hook point (here, two tracepoints) and can pass data out to userspace through a shared memory buffer.

## Why this exists

The default `PollingProcessViewer` ([../event/process.go](../event/process.go)) finds new/exited processes by scanning `/proc` once a second and diffing snapshots. That's simple and portable, but it has real limitations: a process that starts and exits inside a single polling interval is invisible, and there's an inherent latency and CPU cost to a poll loop.

`EBPFProcessViewer` replaces polling with a push model: the kernel tells us, immediately, whenever a process is exec'd or a process exits, via tracepoints — no scanning, no missed short-lived processes.

## The pieces involved

| Layer | File | Role |
|---|---|---|
| eBPF C source | [monitor.c](monitor.c) | Runs in the kernel; hooks two tracepoints and writes an event struct into a ring buffer for each. |
| Code generation | [gen.go](gen.go) | A `go:generate` directive that compiles `monitor.c` and generates Go bindings for it. |
| Generated bindings | `bpf_bpfel.go`, `bpf_bpfeb.go`, `bpf_bpfel.o`, `bpf_bpfeb.o` (git-ignored, not committed) | Go types and loader functions produced by `bpf2go`, `package bpf`. |
| Userspace Go | [../event/process_ebpf.go](../event/process_ebpf.go) | Loads the compiled program, attaches it to the tracepoints, reads events off the ring buffer, and turns them into `event.Event`s. |

Nothing under `bpf/` other than `monitor.c` and `gen.go` is committed to the repo — the generated `.o` and `.go` files are build output, regenerated on demand (see [Building and regenerating](#building-and-regenerating) below).

## monitor.c: the kernel side

### Build constraints

```c
//go:build ignore
// +build ignore
```

This file is C, not Go, but it lives inside a Go module directory. These tags tell the Go toolchain to ignore it when compiling Go packages — it's only ever fed to `clang` (via `bpf2go`), never to `go build` directly.

### License

```c
char __license[] SEC("license") = "GPL";
```

The kernel's eBPF verifier refuses to load certain helper functions (including the ring buffer helpers used here) into programs that don't declare a GPL-compatible license. This is a hard kernel requirement, not a project choice — the `SEC("license")` variable is how a `.o` file communicates its license to the loader.

### The event struct

```c
struct process_event {
    __u32 pid;
    __u32 type;
    char comm[16];
};
```

This is the payload sent from kernel to userspace for every event: the process ID, an event type (`EVENT_STARTED` = 0 or `EVENT_STOPPED` = 1), and `comm` — the kernel's short (16-byte, including the null terminator) name for the executable, e.g. `"busybox"`.

Fixed-size, fixed-layout structs like this are the norm in eBPF: there's no dynamic allocation available, and the struct's layout has to be predictable so userspace can parse the raw bytes it reads back out.

Immediately below it:

```c
struct process_event *unused_process_event __attribute__((unused));
```

This line does nothing at runtime — it exists purely so `bpf2go` can find the type. `bpf2go` generates a matching Go struct for `-type process_event` (see `gen.go`) by reading BTF (BPF Type Format, essentially embedded debug type info) out of the compiled object. BTF is only emitted for types that are actually reachable from some global symbol; `process_event` is otherwise only ever used as a local variable inside `submit_event`, which wouldn't be enough on its own. This unused global pointer makes the type reachable so its BTF gets emitted, without needing the struct anywhere else in the program's logic.

### The ring buffer map

```c
struct {
    __uint(type, BPF_MAP_TYPE_RINGBUF);
    __uint(max_entries, 1 << 16); // 64KB buffer size
} events SEC(".maps");
```

A BPF *map* is the general mechanism eBPF programs use to share state — with userspace, or with each other. `BPF_MAP_TYPE_RINGBUF` is a map type specifically designed for streaming a sequence of variable-length records from kernel to userspace efficiently: it's a single 64KB circular buffer, and the kernel-side reserve/submit helpers plus the userspace reader coordinate over it without needing a syscall per event (userspace is notified via an epoll-compatible fd and reads directly from a shared mmap'd region).

### Submitting an event

```c
static __always_inline int submit_event(__u32 type) {
    struct process_event *e;

    e = bpf_ringbuf_reserve(&events, sizeof(*e), 0);
    if (!e) {
        return 0;
    }

    e->pid = bpf_get_current_pid_tgid() >> 32;
    e->type = type;
    bpf_get_current_comm(&e->comm, sizeof(e->comm));

    bpf_ringbuf_submit(e, 0);
    return 0;
}
```

Shared by both tracepoint handlers below. The reserve/populate/submit pattern is standard for ring buffers: `bpf_ringbuf_reserve` claims `sizeof(struct process_event)` bytes directly inside the buffer (it can fail and return `NULL` if the buffer is full — the check is mandatory, the verifier will reject the program without it), the fields are filled in in place, and `bpf_ringbuf_submit` marks the record ready for userspace to read. Writing straight into buffer-owned memory like this avoids an extra copy.

`bpf_get_current_pid_tgid()` returns a 64-bit value packing two IDs for the calling task: the low 32 bits are the *thread* ID (`tid`), the high 32 bits are the *thread-group* ID (`tgid`) — which is what userspace calls the PID (the ID of the process as a whole, shared by all its threads). Shifting right by 32 extracts the tgid, i.e. the PID a user would recognize.

`bpf_get_current_comm` copies the current task's short name (`comm`, as seen in `/proc/<pid>/comm`, capped at 16 bytes) into the event.

### The two tracepoints

```c
SEC("tracepoint/sched/sched_process_exec")
int handle_exec(void *ctx) {
    return submit_event(EVENT_STARTED);
}
```

The `SEC()` string is how `bpf2go`/libbpf know which kernel hook to attach this function to, and it names a real kernel tracepoint: `sched_process_exec` fires whenever a task calls `execve` and successfully replaces its image — i.e. whenever a new program starts running. This is a strictly better signal for "a process started" than polling `/proc`, since it fires exactly once, exactly when it happens.

```c
SEC("tracepoint/sched/sched_process_exit")
int handle_exit(void *ctx) {
    __u64 pid_tgid = bpf_get_current_pid_tgid();
    __u32 tid = (__u32)pid_tgid;
    __u32 tgid = pid_tgid >> 32;

    if (tid != tgid) {
        return 0;
    }

    return submit_event(EVENT_STOPPED);
}
```

`sched_process_exit` fires once *per thread* that exits, not once per process. A multi-threaded process would therefore generate one exit event per thread if handled naively, making it look like it "stopped" repeatedly. The fix is the `tid != tgid` check: for the thread-group leader (the thread whose tid equals the process's tgid — normally the thread that called `execve` to start the process), tid and tgid are equal; for any other thread in the group, they differ. By only submitting an event when `tid == tgid`, this reports exactly one `STOPPED` event per process, when its leader thread exits.

## gen.go and bpf2go

```go
//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -type process_event Bpf monitor.c -- -O2 -g -Wall
```

Running `go generate ./bpf/` invokes [`bpf2go`](https://github.com/cilium/ebpf), a code generator from the `cilium/ebpf` project. It:

1. Invokes `clang` to compile `monitor.c` to BPF bytecode (`-O2 -g -Wall` are passed straight through to `clang`), producing two object files — `bpf_bpfel.o` (little-endian targets, e.g. x86_64/arm64) and `bpf_bpfeb.o` (big-endian targets).
2. Reads the BTF debug info out of the compiled object to generate matching Go types — this is where `BpfProcessEvent` (from `-type process_event`) comes from, mirroring the C struct field-for-field.
3. Generates a `BpfObjects` struct with a field per program (`HandleExec`, `HandleExit`) and per map (`Events`), plus a `LoadBpfObjects` function to load and verify the compiled bytecode into the kernel.
4. Embeds the compiled `.o` bytes into the generated `.go` files (via `go:embed`), so the final Go binary has no runtime dependency on the `.o` files — they're baked in at compile time.

The `Bpf` argument is a name prefix, giving the generated identifiers their `Bpf...` prefix (`BpfObjects`, `BpfProcessEvent`, etc.) and producing files named `bpf_bpfel.go` / `bpf_bpfeb.go`.

None of this generated output is committed — it's produced fresh by `go generate`, and requires `clang` plus `libbpf-dev` headers (for `<bpf/bpf_helpers.h>`) on the host running it. See [Building and regenerating](#building-and-regenerating).

## process_ebpf.go: the userspace side

### Build tag

```go
//go:build ebpf
```

This file only compiles with `-tags ebpf`, because it imports `package bpf`'s generated bindings, which only exist after `go generate ./bpf/` has run. Everywhere else (including plain `go build`/`go vet` during normal development), [`process_ebpf_stub.go`](../event/process_ebpf_stub.go) provides a same-named `EBPFProcessViewer` that just fails with a clear error if started, so the rest of the module builds cleanly without needing a Linux host or the eBPF toolchain.

### Start(): loading and attaching

```go
func (p *EBPFProcessViewer) Start() {
	if err := rlimit.RemoveMemlock(); err != nil {
		log.Fatalf("failed to remove memlock limit: %v", err)
	}

	objs := bpf.BpfObjects{}
	if err := bpf.LoadBpfObjects(&objs, nil); err != nil {
		log.Fatalf("loading bpf objects failed: %v", err)
	}
	defer objs.Close()
```

- `rlimit.RemoveMemlock()` lifts the process's `RLIMIT_MEMLOCK` (locked-memory) limit. Older kernels charge the memory backing BPF maps and programs against this limit, which defaults low enough that loading a BPF program can fail with a confusing `EPERM`/`ENOMEM`; this is the standard workaround used by essentially every `cilium/ebpf` program. (On kernels with `CONFIG_BPF_JIT` cgroup memory accounting, this is largely a no-op, but it's harmless to always call.)
- `bpf.LoadBpfObjects` (generated by `bpf2go`) loads the embedded compiled bytecode into the kernel — this is the point where the kernel's verifier checks the program (bounds checks, no unbounded loops, valid map accesses, etc.) and rejects it if it isn't provably safe. On success, `objs` holds live kernel handles: `objs.HandleExec` / `objs.HandleExit` (the loaded programs) and `objs.Events` (the ring buffer map).
- `objs.Close()` releases those kernel resources when `Start()` returns.

```go
	execTp, err := link.Tracepoint("sched", "sched_process_exec", objs.HandleExec, nil)
	...
	exitTp, err := link.Tracepoint("sched", "sched_process_exit", objs.HandleExit, nil)
	...
```

`link.Tracepoint` (from `cilium/ebpf/link`) attaches a loaded program to a named kernel tracepoint — here, the same `sched_process_exec` / `sched_process_exit` tracepoints declared via `SEC()` in the C source. This is the step that actually makes the kernel start invoking the program; the returned `link.Link` needs to stay open (its `Close()` detaches it) for the duration of monitoring, which is why both are deferred to the end of `Start()`.

```go
	rd, err := ringbuf.NewReader(objs.Events)
	...
	defer rd.Close()
```

Opens a reader over the `events` ring buffer map. Under the hood this mmaps the ring buffer and sets up an epoll wait, so `rd.Read()` below blocks efficiently until a new record is submitted from the kernel side, rather than polling.

### The read loop

```go
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
```

Each `rd.Read()` blocks until `submit_event` in the C program submits a record, then returns it as raw bytes (`record.RawSample`) — the ring buffer only knows about bytes, not the `process_event` struct shape. `binary.Read` reinterprets those bytes as a `bpf.BpfProcessEvent` (the Go struct `bpf2go` generated from `struct process_event`), using little-endian byte order to match the target architecture. `ringbuf.ErrClosed` is the expected error when `rd.Close()` is called elsewhere (e.g. on shutdown), used here to exit the loop cleanly rather than logging spurious errors.

```go
		comm := string(bytes.TrimRight(int8sToBytes(raw.Comm[:]), "\x00"))

		eventType := "STARTED"
		if raw.Type == ebpfEventStopped {
			eventType = "STOPPED"
		}

		p.handler.HandleEvent(BuildEvent(eventType, int(raw.Pid), comm))
	}
}
```

`raw.Comm` comes through as `[16]int8` — `bpf2go` maps C's `char` faithfully, and `char` is a signed type on the target platform, hence `int8` rather than `byte`/`uint8`. `int8sToBytes` (a small helper in this file) just reinterprets each signed byte as unsigned so it can be handled as a normal Go `[]byte`/string. `bytes.TrimRight(..., "\x00")` strips the trailing NUL padding the kernel leaves in the fixed 16-byte `comm` field.

`raw.Type` is compared against `ebpfEventStopped` to map the numeric `EVENT_STARTED`/`EVENT_STOPPED` constants from the C side (mirrored in Go as `ebpfEventStarted = 0` / `ebpfEventStopped = 1`) back to the same `"STARTED"`/`"STOPPED"` strings the polling viewer uses, via the shared `BuildEvent` helper ([../event/process.go](../event/process.go)) — so downstream code (the plugin pipeline, output handlers) doesn't need to know or care which `ProcessViewer` produced a given event.

## End-to-end event lifecycle

1. Some process execs, or a thread-group leader exits, inside the container `orthanc-observer` shares a PID namespace with.
2. The kernel invokes `handle_exec`/`handle_exit` in `monitor.c` at the corresponding tracepoint.
3. `submit_event` reserves space in the `events` ring buffer, fills in `pid`, `type`, and `comm`, and submits it.
4. The kernel wakes up the blocked `rd.Read()` in `process_ebpf.go`.
5. The raw bytes are decoded into a `BpfProcessEvent`, converted into an `event.Event` via `BuildEvent`, and handed to `p.handler.HandleEvent(...)`.
6. From there it flows through the same path as a polled event: onto the shared channel in `main.go`, through `plugin.RunPipeline` for enrichment, and out through `CompositeEventHandler` to stdout and the log file.

## Building and regenerating

Regenerating the bindings requires a Linux host with `clang` and the libbpf headers:

```
apt-get install clang libbpf-dev
cd client/observer
make generate        # go generate ./bpf/
make build-ebpf       # generate + go build -tags ebpf
```

The Docker build ([../../docker/observer.Dockerfile](../../docker/observer.Dockerfile)) does exactly this inside the build stage, so `docker compose up` (`make docker-up` from the repo root) doesn't require any of this tooling on the host machine — only Docker itself.

At runtime, loading BPF programs and attaching tracepoints needs `CAP_BPF` and `CAP_PERFMON` (kernel 5.8+), granted to the container via `cap_add` in [../../docker/docker-compose.yml](../../docker/docker-compose.yml) (or `privileged: true` on older kernels). The compose file also bind-mounts the host's `/sys/kernel/debug`, since tracepoint attachment reads the tracefs/debugfs event format files, which aren't visible in a container's default fresh sysfs mount otherwise.

## Suggested next steps for learning

A few natural directions to extend this if you want to go deeper:

- Capture more fields — e.g. `ppid`, full command line (would need CO-RE / `bpf_probe_read` against task_struct, or reading `/proc/<pid>/cmdline` from userspace as `BusyboxPlugin` does), or exit code from `sched_process_exit`'s tracepoint arguments (currently unused — `ctx` is untyped `void *`).
- Use `BPF_CORE_READ`/vmlinux.h and real CO-RE (Compile Once – Run Everywhere) to read kernel struct fields directly, rather than relying only on built-in helpers.
- Add a `BPF_MAP_TYPE_HASH` map to filter events in-kernel (e.g. only report processes matching a name), instead of doing all filtering in the plugin pipeline.
- Compare ring buffer (`BPF_MAP_TYPE_RINGBUF`) against the older perf buffer (`BPF_MAP_TYPE_PERF_EVENT_ARRAY`) it superseded, and why the ring buffer's single shared buffer (vs. one per-CPU) simplifies ordering.
