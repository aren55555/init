# SSH access to test VMs, with a dedicated key that lives in 1Password.
# Only the public half is written to disk.

key_pub := "test/out/vm-key.pub"
key_title := "init test VMs"
create := 'op item create --category ssh --title "' + key_title + '" --ssh-generate-key ed25519 >/dev/null && echo "Created 1Password item \"' + key_title + '\"" >&2'
save := 'mkdir -p "$(dirname ' + key_pub + ')" && op item get "' + key_title + '" --fields "public key" > ' + key_pub

[private]
default:
    @just --list ssh

# Print the public key, creating the key in 1Password the first time
key:
    #!/bin/sh
    set -eu
    if [ ! -s {{key_pub}} ]; then
        if ! op item get "{{key_title}}" >/dev/null 2>&1; then
            {{create}}
        fi
        {{save}}
    fi
    cat {{key_pub}}

# Archive the current key in 1Password (after confirming) and generate a new one
rotate:
    #!/bin/sh
    set -eu
    if op item get "{{key_title}}" >/dev/null 2>&1; then
        printf 'Archive the 1Password key "%s" and generate a new one? VMs kept so far will stop accepting it. [y/N] ' "{{key_title}}"
        read -r answer
        case "$answer" in
        [yY]*) op item delete "{{key_title}}" --archive ;;
        *) echo "Kept the current key."; exit 0 ;;
        esac
    fi
    {{create}}
    {{save}}
    echo "Saved the new public key to {{key_pub}}"

# SSH into a test VM (default: the newest) with the test key; a VM started here stops when you exit
shell vm="":
    #!/bin/sh
    set -eu
    vm="{{vm}}"
    [ -n "$vm" ] || vm=$(just vm _newest)
    # Attach to a VM that's already running (e.g. mid-test) as is; only stop VMs started here.
    if ! tart list --source local | awk -v vm="$vm" '$2 == vm && $NF == "running" { found = 1 } END { exit !found }'; then
        tart run --no-graphics --dir=init:{{justfile_directory()}}:ro "$vm" >/dev/null 2>&1 &
        trap 'tart stop "$vm"; wait' EXIT
    fi
    # Record the session on the host: .logs/<vm>-shell-<time>.log
    mkdir -p .logs
    log=".logs/$vm-shell-$(date +%Y%m%d-%H%M%S).log"
    echo "Recording this session to $log" >&2
    script -q "$log" ssh -o IdentityFile={{key_pub}} -o IdentitiesOnly=yes \
        -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -o LogLevel=ERROR \
        -o ConnectionAttempts=30 admin@$(tart ip --wait 120 "$vm")
