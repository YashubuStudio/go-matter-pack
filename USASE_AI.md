# go-matter-pack AI向け完全ガイド

このドキュメントは、`go-matter-pack` を他プロジェクトへ組み込む際に、AI が**コピペのみで全体像を理解し、すぐ使える**レベルの情報をまとめたものです。API は開発中で変更される可能性があるため、最新の README と `doc/` 配下も合わせて参照してください。

## 1. このリポジトリの提供物

- **Go ライブラリ**: Matter デバイスの発見・コミッショニング等を扱うライブラリ。
  - モジュールパス: `github.com/cybergarage/go-matter`
- **CLI**: `matterctl`（Matter デバイス操作用）
  - エントリーポイント: `cmd/matterctl`

> 進捗状況は README の「Packages」テーブルを参照。

## 2. 主要ディレクトリと役割

- `matter/` : ライブラリの主要 API（Commissioner 等）
- `matter/encoding/` : ペアリングコードや QR などのエンコード/デコード
- `internal/` : CLI/アプリ向けの内部実装（Store など）
- `cmd/matterctl/` : CLI 実装
- `doc/` : CLI やインストールのドキュメント

## 3. インストール/ビルド

### 3.1 Go モジュールとして利用
```bash
go get github.com/cybergarage/go-matter
```

### 3.2 CLI ビルド
```bash
# 既定名

go build -o matterctl ./cmd/matterctl

# パッケージ名を維持したい場合

go build -o go-matter-pack ./cmd/matterctl
```

### 3.3 ローカルインストール
```bash
sudo install -m 0755 matterctl /usr/local/bin/matterctl
```

> systemd で常駐させる場合は `doc/INSTALL.md` を参照。

## 4. Go ライブラリ: 主要 API と最小実装例

### 4.1 デバイス探索（Discovery）
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

### 4.2 手動ペアリングコード解析
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

    log.Printf("VID=%d PID=%d Discriminator=%d Passcode=%d",
        pairingCode.VendorID(),
        pairingCode.ProductID(),
        pairingCode.Discriminator(),
        pairingCode.Passcode(),
    )
}
```

### 4.3 コミッショニング（ペアリング）
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

## 5. CLI `matterctl` の基本

### 5.1 主要サブコマンド
- スキャン: `matterctl scan`
- ペアリング (手動コード): `matterctl pairing code <node ID> <pairing code>`
- ペアリング (Wi-Fi 付き): `matterctl pairing code-wifi <node ID> <pairing code> <WIFI SSID> <WIFI password>`

### 5.2 出力形式と共通フラグ
- `--format` : `table` / `json` / `csv`
- `--debug` : デバッグ出力
- `--verbose` : 詳細ログ

### 5.3 ヘルプ/ドキュメント
```bash
matterctl help
matterctl doc
```

> 完全なヘルプは `doc/matterctl.md` を参照。

## 6. 状態保存（commission.json）と再利用

`matterctl` および内部 API は、コミッショニング情報を永続化する設計になっています。

### 6.1 保存場所
- `internal/app` のヘルパーは XDG の state ディレクトリを利用
  - `$XDG_STATE_HOME` または `~/.local/state`
- 既定ファイル名として `commission.json` を扱う実装例が用意されています。

### 6.2 保存内容
`commission.json` には以下が保存されます:

- **payload**
  - QR または手動ペアリングコード情報（VID/PID/Discriminator/Passcode など）
- **bundle**
  - Operational 証明書/鍵や IPK
- **result**
  - 直近のコミッショニング結果（Node ID 等）

### 6.3 セキュリティ注意
- 鍵素材やパスコードが含まれるため、権限管理が必須。
- `JSONFileStore` は `0600` 権限で保存。

## 7. 内部ストア API の概要（保存と読み込み）

`internal/store.Store` が保存 I/F。既定実装は `JSONFileStore` です。

```go
package main

import (
    "context"
    "log"
    "path/filepath"

    "github.com/cybergarage/go-matter/internal/app"
    "github.com/cybergarage/go-matter/internal/commission"
    "github.com/cybergarage/go-matter/internal/store"
)

func main() {
    stateDir := app.StateDir("go-matter-pack")
    statePath := filepath.Join(stateDir, "commission.json")
    st := store.NewJSONFileStore(statePath)

    state, err := commission.LoadState(context.Background(), st)
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("loaded payload: %+v", state.Payload)
}
```

## 8. 典型的な統合フロー

1. **探索**: `Commissioner.Discover` でデバイスを検出
2. **オンボーディング情報の取得**: QR または手動コードを用意
3. **コミッショニング**: `Commissioner.Commission` でペアリング
4. **状態保存**: `commission.json` に保存し再起動後の再利用を想定

## 9. 参考リンク

- `README.md` / `README.ja.md`
- `doc/INSTALL.md`
- `doc/matterctl.md`

---

このファイルをプロンプトに貼り付ければ、AI が `go-matter-pack` の構成・CLI・API の概念を即座に把握できるよう意図されています。
