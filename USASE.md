# 使い方まとめ (go-matter-pack)

## 概要
`go-matter-pack` は Matter アプリケーション/デバイスを Go で扱うためのライブラリと、操作用 CLI (`matterctl`) を提供します。プロジェクトは開発途中のため API 変更があり得ます。参照元は README と付属ドキュメントです。 

## パッケージとしての簡単な使い方 (Go モジュール)

### 追加
```bash
go get github.com/cybergarage/go-matter
```

### デバイス検索 (Discover)
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

## CLI (matterctl) の基本

### ビルド/インストール
- ビルド: `go build -o matterctl ./cmd/matterctl`
- パッケージ名として作る場合: `go build -o go-matter-pack ./cmd/matterctl`
- ローカルインストール例: `sudo install -m 0755 matterctl /usr/local/bin/matterctl`

詳細は `doc/INSTALL.md` を参照してください。

### 主要サブコマンド
- `matterctl scan` : Matter デバイスのスキャン
- `matterctl pairing` : Matter デバイスのペアリング
  - `matterctl pairing code <node ID> <pairing code>`
  - `matterctl pairing code-wifi <node ID> <pairing code> <WIFI SSID> <WIFI password>`

詳細なフラグや出力形式は `doc/matterctl.md` のヘルプを参照してください。

## デモ動作を行うためのコード一覧

### 1. 手動ペアリングコードの解析
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

### 2. デバイススキャン (Commissioner)
```go
package main

import (
    "context"
    "log"

    "github.com/cybergarage/go-matter/matter"
)

func main() {
    commissioner := matter.NewCommissioner()

    devices, err := commissioner.Discover(context.Background(), matter.NewQuery())
    if err != nil {
        log.Fatal(err)
    }

    for _, device := range devices {
        log.Printf("device: %+v", device)
    }
}
```

### 3. ペアリング (Commission)
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

## 参考
- README (日本語/英語)
- `doc/INSTALL.md`
- `doc/matterctl.md`
