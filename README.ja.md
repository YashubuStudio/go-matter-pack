![](https://img.shields.io/badge/status-Work%20In%20Progress-8A2BE2)
![](https://workers-hub.zoom.us/j/89428436853?pwd=Qm41UHlJNW1LazN3RFVzV1dwM09udz09&from=addon)
![GitHub tag (latest SemVer)](https://img.shields.io/github/v/tag/cybergarage/go-matter)
[![test](https://github.com/cybergarage/go-matter/actions/workflows/make.yml/badge.svg)](https://github.com/cybergarage/go-matter/actions/workflows/make.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/cybergarage/go-matter.svg)](https://pkg.go.dev/github.com/cybergarage/go-matter)
 [![Go Report Card](https://img.shields.io/badge/go%20report-A%2B-brightgreen)](https://goreportcard.com/report/github.com/cybergarage/go-matter) 
 [![codecov](https://codecov.io/gh/cybergarage/go-matter/graph/badge.svg?token=7Y64KS92VD)](https://codecov.io/gh/cybergarage/go-matter)

# go-matter

Matter はスマートホームおよび IoT（Internet of Things）デバイス向けのオープンソース接続標準です。
`go-matter` は Matter アプリケーションやデバイスを開発するための Go ライブラリです。

**注記:** 🌱 このプロジェクトは余暇ベースの趣味プロジェクトのため、進捗は緩やかで変更は不定期になる場合があります。ご了承ください 🙂

**重要:** このリポジトリは [cybergarage/go-matter](https://github.com/cybergarage/go-matter) のフォークです。

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

## パッケージ利用

`go-matter` は Go モジュールとしてアプリケーションに組み込めます。

```bash
go get github.com/cybergarage/go-matter
```

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

> **注記**
> 本プロジェクトは進行中のため、API は今後変更される可能性があります。


## 参考資料

- [Matter](https://buildwithmatter.com/)
    - [Matter 1.5 Standard Namespace Specification](https://csa-iot.org/developer-resource/specifications-download-request/)
    - [Matter 1.5 Device Library Specification](https://csa-iot.org/developer-resource/specifications-download-request/)
    - [Matter 1.5 Core Specification](https://csa-iot.org/developer-resource/specifications-download-request/)
    - [Matter 1.5 Application Cluster Specification](https://csa-iot.org/developer-resource/specifications-download-request/)
