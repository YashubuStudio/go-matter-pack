# go-matter-pack Quick Start (AI-oriented)

## What this repository provides
- A Go library for Matter devices (`github.com/cybergarage/go-matter`).
- A CLI called `matterctl` for scanning and pairing devices.

## Install / Build
### Library dependency
```bash
go get github.com/cybergarage/go-matter
```

### Build the CLI
```bash
go build -o matterctl ./cmd/matterctl
# or packaged name
# go build -o go-matter-pack ./cmd/matterctl
```

## Minimal usage examples

### Discover devices
```go
package main

import (
    "context"
    "log"

    "github.com/cybergarage/go-matter/matter"
)

func main() {
    commissioner := matter.NewCommissioner()
    query := matter.NewQuery()

    devices, err := commissioner.Discover(context.Background(), query)
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("discovered %d devices", len(devices))
}
```

### Parse a manual pairing code
```go
package main

import (
    "log"

    "github.com/cybergarage/go-matter/matter/encoding"
)

func main() {
    pairingCode, err := encoding.NewPairingCodeFromString("34970112332")
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("VID=%d PID=%d Discriminator=%d Passcode=%d", pairingCode.VendorID(), pairingCode.ProductID(), pairingCode.Discriminator(), pairingCode.Passcode())
}
```

### Commission (pair) a device
```go
package main

import (
    "context"
    "log"

    "github.com/cybergarage/go-matter/matter"
    "github.com/cybergarage/go-matter/matter/encoding"
)

func main() {
    commissioner := matter.NewCommissioner()

    payload, err := encoding.NewPairingCodeFromString("34970112332")
    if err != nil {
        log.Fatal(err)
    }

    commissionee, err := commissioner.Commission(context.Background(), payload)
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("commissioned: %+v", commissionee)
}
```

## CLI cheat sheet
- Scan for devices: `matterctl scan`
- Pair with a manual code: `matterctl pairing code <node ID> <pairing code>`
- Pair with Wi-Fi credentials: `matterctl pairing code-wifi <node ID> <pairing code> <WIFI SSID> <WIFI password>`

## References
- `doc/INSTALL.md` for packaging and systemd setup.
- `doc/matterctl.md` for detailed CLI flags and help output.
