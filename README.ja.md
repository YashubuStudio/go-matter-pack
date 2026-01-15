![](https://img.shields.io/badge/status-Work%20In%20Progress-8A2BE2)
![](https://workers-hub.zoom.us/j/89428436853?pwd=Qm41UHlJNW1LazN3RFVzV1dwM09udz09&from=addon)
![GitHub tag (latest SemVer)](https://img.shields.io/github/v/tag/cybergarage/go-matter)
[![test](https://github.com/YashubuStudio/go-matter-pack/actions/workflows/make.yml/badge.svg)](https://github.com/YashubuStudio/go-matter-pack/actions/workflows/make.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/YashubuStudio/go-matter-pack.svg)](https://pkg.go.dev/github.com/YashubuStudio/go-matter-pack)
 [![Go Report Card](https://img.shields.io/badge/go%20report-A%2B-brightgreen)](https://goreportcard.com/report/github.com/YashubuStudio/go-matter-pack) 
 [![codecov](https://codecov.io/gh/cybergarage/go-matter/graph/badge.svg?token=7Y64KS92VD)](https://codecov.io/gh/cybergarage/go-matter)

# go-matter

Matter はスマートホームおよび IoT（Internet of Things）デバイス向けのオープンソース接続標準です。
`go-matter` は Matter アプリケーションやデバイスを開発するための Go ライブラリです。

**注記:** 🌱 このプロジェクトは余暇ベースの趣味プロジェクトのため、進捗は緩やかで変更は不定期になる場合があります。ご了承ください 🙂

**重要:** このリポジトリは [cybergarage/go-matter](https://github.com/YashubuStudio/go-matter-pack) のフォークです。

### 進捗概要

#### パッケージ

| カテゴリ | パッケージ | ステータス | 説明 |
|----------|---------|--------|-------------|
| Discovery | `ble.btp` | ✅ 検証中 | BLE トランスポートプロトコル実装 |
|           | `mdns` | ✅ 検証中 | mDNS (Multicast DNS) サービスディスカバリ |
| Commissioning |`pase` | 🚧 作業中 | パスコード認証セッション確立 (PASE) 実装 |
| Encoding | `encoding.base38` | ✅ 実装済み | Base38 エンコード/デコード |
|          | `encoding.qr` | ✅ 実装済み | QR コード生成 |
|          | `encoding.pairing` | ✅ 実装済み | 手動ペアリングコード処理 |
|          | `encoding.tlv` | 🚧 作業中 | TLV (Tag-Length-Value) エンコード |

#### 関連プロジェクト

| プロジェクト | ステータス | 説明 |
|---------|--------|-------------|
| [go-ble](https://github.com/cybergarage/go-ble) | 🚧 作業中 | Bluetooth Low Energy (BLE) 通信の Go パッケージ |
| [go-mdns](https://github.com/cybergarage/go-mdns) | 🚧 作業中 | mDNS (Multicast DNS) サービスディスカバリの Go パッケージ |


# ユーザーガイド

- Operation
  - [matterctl](doc/matterctl.md)
- インストール
  - [INSTALL](doc/INSTALL.md)

## クイックスタート（検索・探索・Commissioning・OnOff）

> **PowerShell の注意**: `matterctl.exe` を同じディレクトリで実行する場合は、`.\matterctl.exe` のように `.\` を付けて実行してください。

### 1) 検索・探索（Scan）

```
.\matterctl.exe scan --format table
```

- `--format` は `table`/`json`/`csv` を指定できます。
- 反応がない場合はデバイスが Commissionable モードであるか（初期化直後など）を確認してください。

### 2) Commissioning（セットアップ）

```
.\matterctl.exe setup commission --code <11桁の手動ペアリングコード> --node-id <node-id>
```

#### 直接アドレスを指定する場合

```
.\matterctl.exe setup commission --code <11桁の手動ペアリングコード> --node-id <node-id> --address <IP[:PORT]>
```

- 実行後、commissioning 結果は `commission.json` に保存されます。
  - 既定の保存先は OS の XDG state ディレクトリです（例: `C:\Users\<user>\.local\state\go-matter-pack\commission.json`）。
  - 保存先を変更したい場合は `--state-dir` を使用してください。
- コードの登録のみ行いたい場合は `--import-only` を付けてください。

### 3) OnOff 操作について

現時点の `matterctl` には `onoff` コマンドは実装されていません。そのため、OnOff の操作を行いたい場合は以下のいずれかを検討してください。

- Matter 対応の別 CLI（例: チップセットの提供するコントローラ）で OnOff クラスタのコマンドを送る。
- `go-matter-pack` を利用して独自のコントローラを実装する。

## パッケージ利用

`go-matter` は Go モジュールとしてアプリケーションに組み込めます。

```bash
go get github.com/YashubuStudio/go-matter-pack
```

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

> **注記**
> 本プロジェクトは進行中のため、API は今後変更される可能性があります。


## 参考資料

- [Matter](https://buildwithmatter.com/)
    - [Matter 1.5 Standard Namespace Specification](https://csa-iot.org/developer-resource/specifications-download-request/)
    - [Matter 1.5 Device Library Specification](https://csa-iot.org/developer-resource/specifications-download-request/)
    - [Matter 1.5 Core Specification](https://csa-iot.org/developer-resource/specifications-download-request/)
    - [Matter 1.5 Application Cluster Specification](https://csa-iot.org/developer-resource/specifications-download-request/)
