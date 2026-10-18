# Test VMs: `just vm` lists the recipes
mod vm

# SSH into test VMs: `just ssh` lists the recipes
mod ssh

[private]
default:
    @just --list --list-submodules

# Run the Go unit tests
test:
    go test ./...

# Vet Go and syntax-check the shell files
lint:
    go vet ./...
    sh -n bootstrap.sh
    zsh -n dotfiles/zsh/.zshrc

# Show what the installer would do on this machine, without changing anything
dry-run:
    go run ./cmd/init --dry-run --yes
