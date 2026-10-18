# Test VMs: fresh macOS VMs (Tart) provisioned with bootstrap.sh.

[private]
default:
    @just --list vm

# Bootstrap a fresh VM and check the result (flags: -keep, -remote)
test *flags:
    go run ./cmd/vmtest {{flags}}

# Bootstrap a fresh VM, check it, then open a shell in it (kept; `just vm clean` removes it)
try:
    #!/bin/sh
    set -eu
    key=$(just ssh key)
    name=init-test-$(date +%Y%m%d-%H%M%S)
    go run ./cmd/vmtest -keep -name "$name" -authorized-key "$key" ||
        echo "Some checks failed; opening a shell anyway so you can look around."
    just ssh shell "$name"

# Open a test VM's desktop (default: the newest; login: admin / admin)
gui vm="":
    #!/bin/sh
    set -eu
    vm="{{vm}}"
    [ -n "$vm" ] || vm=$(just vm _newest)
    tart run --dir=init:{{justfile_directory()}}:ro "$vm"

# List VMs
list:
    tart list --source local

# Stop and delete every test VM (keeps mac-base)
clean:
    #!/bin/sh
    for vm in $(tart list --source local | awk '$2 ~ /^init-test-/ { print $2 }'); do
        tart stop "$vm" 2>/dev/null || true
        tart delete "$vm" && echo "deleted $vm"
    done

# Print the newest test VM's name, or fail if there are none
[private]
_newest:
    #!/bin/sh
    vm=$(tart list --source local | awk '$2 ~ /^init-test-/ { print $2 }' | sort | tail -n 1)
    if [ -z "$vm" ]; then
        echo "No test VMs. Make one with \`just vm try\`." >&2
        exit 1
    fi
    echo "Using $vm" >&2
    echo "$vm"
