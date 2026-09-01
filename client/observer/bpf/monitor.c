//go:build ignore
// +build ignore

#include <linux/bpf.h>
#include <bpf/bpf_helpers.h>

char __license[] SEC("license") = "GPL";

#define EVENT_STARTED 0
#define EVENT_STOPPED 1

// Define the event data structure passed to Go
struct process_event {
    __u32 pid;
    __u32 type;
    char comm[16];
};

// Define the Ring Buffer map
struct {
    __uint(type, BPF_MAP_TYPE_RINGBUF);
    __uint(max_entries, 1 << 16); // 64KB buffer size
} events SEC(".maps");

static __always_inline int submit_event(__u32 type) {
    struct process_event *e;

    // Reserve space in the ring buffer
    e = bpf_ringbuf_reserve(&events, sizeof(*e), 0);
    if (!e) {
        return 0;
    }

    // Populate data using helper functions
    e->pid = bpf_get_current_pid_tgid() >> 32; // Extract the user-space PID
    e->type = type;
    bpf_get_current_comm(&e->comm, sizeof(e->comm)); // Extract executable name

    // Submit the event back to user space (Go)
    bpf_ringbuf_submit(e, 0);
    return 0;
}

// Tracepoint triggered whenever a new process is executed
SEC("tracepoint/sched/sched_process_exec")
int handle_exec(void *ctx) {
    return submit_event(EVENT_STARTED);
}

// Tracepoint triggered whenever a task exits. sched_process_exit fires once
// per thread, so only report the event when the exiting thread is the
// thread-group leader - otherwise a multi-threaded process would look like
// it "stopped" every time any one of its threads exited.
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
