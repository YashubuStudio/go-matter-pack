# 使い方まとめ (go-matter-pack)

## 概要
`go-matter-pack` は Matter アプリケーション/デバイスを Go で扱うためのライブラリと、操作用 CLI (`matterctl`) を提供します。プロジェクトは開発途中のため API 変更があり得ます。参照元は README と付属ドキュメントです。 

## パッケージとしての簡単な使い方 (Go モジュール)

### 追加
```bash
go get github.com/YashubuStudio/go-matter-pack
```

### デバイス検索 (Discover)
```go
package main

import (
    "context"
    "log"

    "github.com/YashubuStudio/go-matter-pack/matter"
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

### 設定ファイル (matterctl.yaml)
- `matterctl` は設定ファイルを参照し、存在しない場合は **実行ファイルと同階層** に `matterctl.yaml` を自動生成します。
- 別の場所に置く場合は `--config` でパスを指定してください。
- discovery を使う場合は `enable-ble` または `enable-mdns` を `true` に設定します。
- 設定ファイルはデフォルト値のまま利用し、必要な指定はコマンドラインで行います。コマンドラインの指定が設定ファイルより優先されます。

例: `matterctl.yaml`
```yaml
# matterctl configuration
# Set enable-ble or enable-mdns to true for discovery.
format: table
verbose: false
debug: false
enable-ble: false
enable-mdns: true
```

### バイナリ運用の考え方 (永続化/配置)
- 運用では `matterctl` バイナリを固定のパスに配置し、バージョンや再起動手順を明確にします。
  - 例: `/usr/local/bin/matterctl` を標準の配置先として運用。
- systemd などでサービス化する場合は、`doc/INSTALL.md` の例に従い、実行ユーザーの権限と状態保存先 (後述) を合わせて管理します。
- バイナリ更新時は保存状態 (`commission.json` など) を保持したまま差し替え、状態ファイルの互換性に注意します。

### 主要サブコマンド
- `matterctl scan` : Matter デバイスのスキャン
- `matterctl pairing` : Matter デバイスのペアリング
  - `matterctl pairing code <node ID> <pairing code>`
  - `matterctl pairing code-wifi <node ID> <pairing code> <WIFI SSID> <WIFI password>`
- `matterctl setup commission` : オンボーディング情報を保存し、必要に応じてコミッショニングを実行
  - `matterctl setup commission --qr <payload> --node-id <node ID>`
  - `matterctl setup commission --code <pairing code> --node-id <node ID>`
  - `--import-only` を付けると保存のみを実行
  - `--state-dir` で保存ディレクトリを明示指定 (省略時は XDG の state ディレクトリを利用)

### IP で登録する手順 (オンネットワーク)
1. `matterctl.yaml` はデフォルト値のまま利用します (必要な指定はコマンドラインで行います)。
2. コミッショニング時に `--address` を指定して対象 IP に直接接続します。
   - IP だけ指定した場合はデフォルトポート (5540) を利用します。

```bash
# QR ペイロードを使う場合
matterctl setup commission --qr <payload> --node-id <node ID> --address <ip>

# 手動ペアリングコードを使う場合 (IP:ポート指定も可)
matterctl setup commission --code <pairing code> --node-id <node ID> --address <ip:port>
```

3. 保存だけを行う場合は `--import-only` を付けます (IP 指定の有無に関係なく利用可能)。

詳細なフラグや出力形式は `doc/matterctl.md` のヘルプを参照してください。

## Matter の永続化/保存設計 (バイナリと接続情報)

### 1. 状態保存のストレージ設計
- `internal/store.Store` は状態保存用の最小インターフェースです。
- 既定実装として `JSONFileStore` があり、指定パスに JSON を保存します。
- CLI (`matterctl setup commission`) では、`--state-dir` を指定しない場合に XDG の state ディレクトリ (`$XDG_STATE_HOME` または `~/.local/state`) を用い、`<state-dir>/commission.json` に保存します。

#### 保存されるデータ
`commission.json` には以下の情報が保存されます。

- **payload** (オンボーディング情報)
  - QR または手動ペアリングコードの情報
  - Vendor ID / Product ID / Discriminator / Passcode など
- **bundle** (運用接続用の証明書/鍵)
  - Operational Certificate / Intermediate / Root
  - Operational Key / IPK など
- **result** (直近のコミッショニング結果)
  - Node ID / Vendor ID / Product ID / Device 表記など

### 2. 接続情報 (運用フェーズ) の保存と再利用
- コミッショニング後に得られる運用証明書や鍵 (Bundle) を `Store` に保存しておくことで、再起動後でも運用セッションを復元できる設計です。
- 保存済みの `bundle` を読み出し、Matter の運用コントローラに渡して再接続する運用を想定しています。

### 3. 保存ポリシーとセキュリティ
- `commission.json` には鍵素材やパスコードなど秘匿情報が含まれます。アクセス権を適切に制御してください。
- `JSONFileStore` は `0600` で保存します。プロセスの実行ユーザーと権限管理を前提にした運用を推奨します。
- バックアップを行う場合は暗号化や安全な保管経路を設けてください。

### 4. 実装例 (アプリケーションからの保存/読み込み)
```go
package main

import (
    "context"
    "log"
    "path/filepath"

    "github.com/YashubuStudio/go-matter-pack/internal/app"
    "github.com/YashubuStudio/go-matter-pack/internal/commission"
    "github.com/YashubuStudio/go-matter-pack/internal/store"
)

func main() {
    stateDir := app.StateDir("go-matter-pack")
    statePath := filepath.Join(stateDir, "commission.json")
    st := store.NewJSONFileStore(statePath)

    // 読み込み
    state, err := commission.LoadState(context.Background(), st)
    if err != nil {
        log.Fatal(err)
    }
    log.Printf("loaded payload: %+v", state.Payload)
}
```

## デモ動作を行うためのコード一覧

### 1. 手動ペアリングコードの解析
```go
package main

import (
    "log"

    "github.com/YashubuStudio/go-matter-pack/matter/encoding"
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

    "github.com/YashubuStudio/go-matter-pack/matter"
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

    "github.com/YashubuStudio/go-matter-pack/matter"
    "github.com/YashubuStudio/go-matter-pack/matter/encoding"
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
