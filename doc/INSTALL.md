# Installation & Packaging

This document describes how to build, install, and package the `go-matter-pack` CLI binary, plus a sample `systemd` unit for running it as a service.

> **Note**
> The CLI entry point is implemented under `cmd/matterctl`. You can keep the binary name as `matterctl` or rename it to `go-matter-pack` during packaging.

## Build the binary

```bash
# Build with the default name

go build -o matterctl ./cmd/matterctl

# Or build with the packaged name

go build -o go-matter-pack ./cmd/matterctl
```

## Install locally

```bash
# Install into /usr/local/bin

sudo install -m 0755 matterctl /usr/local/bin/matterctl
# or
sudo install -m 0755 go-matter-pack /usr/local/bin/go-matter-pack
```

## Create a tar.gz package

```bash
VERSION="$(git describe --tags --always)"
BIN_NAME=go-matter-pack
OUT_DIR="dist/${BIN_NAME}-${VERSION}-linux-amd64"

mkdir -p "${OUT_DIR}"
cp "${BIN_NAME}" "${OUT_DIR}/"
cp LICENSE "${OUT_DIR}/"
cp README.md "${OUT_DIR}/"
cp doc/INSTALL.md "${OUT_DIR}/"

tar -C dist -czf "${BIN_NAME}-${VERSION}-linux-amd64.tar.gz" "${BIN_NAME}-${VERSION}-linux-amd64"
```

## systemd unit (optional)

A sample unit file is available at `doc/systemd/go-matter-pack.service`.

1. Copy the unit file:

   ```bash
   sudo install -m 0644 doc/systemd/go-matter-pack.service /etc/systemd/system/go-matter-pack.service
   ```

2. Adjust the `ExecStart` command to the desired mode/subcommand.

3. Enable and start:

   ```bash
   sudo systemctl daemon-reload
   sudo systemctl enable --now go-matter-pack.service
   ```

The unit uses `DynamicUser=yes`, `StateDirectory=go-matter-pack`, and sets `XDG_STATE_HOME=/var/lib` so state is stored under `/var/lib/go-matter-pack/` by default.
