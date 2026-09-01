# orthanc

`orthanc` is a small process-monitoring sidecar, built as a learning project for **Go** and **eBPF**. It watches another container's processes start and stop, runs the observed events through a pluggable enrichment pipeline, and logs the results.

There's no grand product ambition here — the goal is to have a real, working reason to learn:

- idiomatic Go project structure, interfaces, build tags, and channels/goroutines
- how eBPF programs are written, compiled, loaded, and talked to from userspace
- how a sidecar container observes another container via a shared PID namespace

## How it works

```
┌─────────────────────────┐        ┌───────────────────────────────────┐
│  observer-test-app       │        │  orthanc-observer (sidecar)        │
│  (the container being    │◄───────┤  shares the app's PID namespace    │
│   watched)                │  pid:  │                                     │
└─────────────────────────┘  ns     │  ProcessViewer                     │
                                     │   ├─ PollingProcessViewer (default)│
                                     │   │   polls /proc every 1s         │
                                     │   └─ EBPFProcessViewer (opt-in)    │
                                     │       kernel tracepoints via eBPF  │
                                     │              │                     │
                                     │              ▼                     │
                                     │        event.Event channel         │
                                     │              │                     │
                                     │              ▼                     │
                                     │        plugin.RunPipeline          │
                                     │      (enriches event.Evidence)     │
                                     │              │                     │
                                     │              ▼                     │
                                     │      CompositeEventHandler         │
                                     │       ├─ stdout (JSON)             │
                                     │       └─ /var/log/observer.log     │
                                     └───────────────────────────────────┘
```

`orthanc-observer` runs as a Docker Compose sidecar that shares the PID namespace of the container it's watching (`pid: "service:observer-test-app"` in [docker/docker-compose.yml](docker/docker-compose.yml)). It detects processes starting and stopping inside that shared namespace, turns each into an `event.Event`, and reports it.

### Process viewers

Detecting process start/stop is pluggable behind the `event.ProcessViewer` interface ([client/observer/event/viewer.go](client/observer/event/viewer.go)), with two interchangeable implementations:

- **`PollingProcessViewer`** ([client/observer/event/process.go](client/observer/event/process.go)) — the portable default. Every second it snapshots `/proc/<pid>/exe` for every PID and diffs the snapshot against the previous one to find processes that appeared or disappeared. Works anywhere Go runs; no special privileges needed.
- **`EBPFProcessViewer`** ([client/observer/event/process_ebpf.go](client/observer/event/process_ebpf.go)) — a Linux-only implementation that attaches to the kernel's `sched_process_exec` / `sched_process_exit` tracepoints via an eBPF program and streams events out of a ring buffer as they happen, instead of polling. See **[client/observer/bpf/EBPF.md](client/observer/bpf/EBPF.md)** for the full write-up of how this works, from the C source to the Go bindings.

Which one runs is chosen at startup in [client/observer/main.go](client/observer/main.go) based on `config.UseEBPF()`, which just checks the `OBSERVER_EBPF=1` environment variable ([client/observer/config/config.go](client/observer/config/config.go)).

The eBPF viewer is also gated behind a Go build tag (`-tags ebpf`), since it depends on generated bindings that only exist after running `go generate` on a Linux host with `clang`/`libbpf-dev` installed. Everywhere else, [process_ebpf_stub.go](client/observer/event/process_ebpf_stub.go) provides a stand-in that fails loudly if selected, so the rest of the module still builds cleanly on any platform (e.g. developing on macOS).

### Event pipeline

Whichever viewer is active, every process start/stop becomes an `event.Event` ([client/observer/event/event.go](client/observer/event/event.go)) with a timestamp, PID, process path, and event type (`STARTED`/`STOPPED`, or `SNAPSHOT` for the polling viewer's initial baseline).

`main.go` reads events off a channel and, before handing them to the output handlers, runs each one through `plugin.RunPipeline` ([client/observer/plugin/pipeline.go](client/observer/plugin/pipeline.go)). Each registered `Plugin` ([client/observer/plugin/plugin.go](client/observer/plugin/plugin.go)) can inspect an event and attach `Evidence` — arbitrary key/value facts about the process. Today there's one plugin, `BusyboxPlugin` ([client/observer/plugin/busybox.go](client/observer/plugin/busybox.go)), which reads `/proc/<pid>/cmdline` for any process ending in `/busybox` and records the command and arguments it was invoked with. The pipeline is designed so more plugins can be added without touching the core event flow.

Finished events are handed to a `CompositeEventHandler` ([client/observer/event/handler.go](client/observer/event/handler.go)) which fans them out to:

- `SystemOutEventHandler` — pretty-prints each event as JSON to stdout
- `FileOutEventHandler` — appends each event as JSON to `/var/log/observer.log`

## Running it

### Docker Compose (recommended)

```
make docker-up    # builds and starts observer-test-app + orthanc-observer
make docker-down  # tears it down
```

This builds `orthanc-observer` with the `ebpf` build tag baked in (see [docker/observer.Dockerfile](docker/observer.Dockerfile)) and runs it with `OBSERVER_EBPF=1`, attached to `observer-test-app`'s PID namespace. Loading BPF programs and attaching tracepoints requires the `BPF` and `PERFMON` capabilities (kernel 5.8+), granted in [docker/docker-compose.yml](docker/docker-compose.yml).

### Building locally

```
cd client/observer
make build         # portable build, PollingProcessViewer only
make build-ebpf     # Linux only: go generate + build -tags ebpf
```

`make build-ebpf` requires `clang` and `libbpf-dev` on the build host to compile the eBPF C source into Go bindings (see [client/observer/bpf/gen.go](client/observer/bpf/gen.go)).

## Project layout

```
client/observer/
├── main.go                 wires everything together and starts the ProcessViewer
├── config/                 environment-driven configuration (OBSERVER_EBPF, log path)
├── event/                  Event type, ProcessViewer implementations, output handlers
├── plugin/                 pluggable per-event enrichment pipeline
└── bpf/                    eBPF C source + generated Go bindings (see EBPF.md)
docker/                     Dockerfiles and the Compose stack used to run the sidecar
```
