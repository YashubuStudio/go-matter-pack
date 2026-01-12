# 計画リスト

 [〇] 1. プロジェクト方針の確定（実行モード/保存先/運用方針）

 [〇] 2. 内部モジュール分割（app/store/log）

 [〇] 3. Matter運用接続の抽象化（Controllerインターフェース）

 [〇] 4. ブリッジ子端末の列挙（PartsList→UniqueID/NodeLabel/Reachable）

 [ ] 5. 安定選択の実現（UniqueID主キー＋ラベル固定）

 [ ] 6. Lock/Unlock操作の実装（Door Lock Cluster）

 [ ] 7. 電池・到達性の定期取得（poll/watch）

 [ ] 8. コミッショニング対応（importまたはPASE/CASE実装）

 [ ] 9. パッケージ運用整備（systemd/INSTALL）

※この順番通りに行い、製作開始前に状態を確認するようにしてください。
※チェック：[◆] 製作実行：[★] 制作完了：[〇] 未着手：[ ]

※製作計画に関しては原則として編集を行わないものとする（チェックなどを除く）。変更履歴や変更における要点などはCheck-list.md末尾に必ず追加し、引き継ぐこと。Check-list.mdは最新のものは必ず読み込むこと。

 [〇] 1. プロジェクト方針の確定（実行モード/保存先/運用方針）

 [〇] 2. 内部モジュール分割（app/store/log）

 [〇] 3. Matter運用接続の抽象化（Controllerインターフェース）

 [〇] 4. ブリッジ子端末の列挙（PartsList→UniqueID/NodeLabel/Reachable）

 [ ] 5. 安定選択の実現（UniqueID主キー＋ラベル固定）

 [ ] 6. Lock/Unlock操作の実装（Door Lock Cluster）

 [ ] 7. 電池・到達性の定期取得（poll/watch）

 [ ] 8. コミッショニング対応（importまたはPASE/CASE実装）

 [ ] 9. パッケージ運用整備（systemd/INSTALL）

※この順番通りに行い、製作開始前に状態を確認するようにしてください。
※チェック：[◆] 製作実行：[★] 制作完了：[〇] 未着手：[ ]

※製作計画に関しては原則として編集を行わないものとする（チェックなどを除く）。変更履歴や変更における要点などは

# 追加仕様（あなたの要件を “壊れない形” に落とす）

以下を満たすと「番号変更/再登録で困る」問題がほぼ潰せます。

### A. 永続化（超重要）

* **Fabric/証明書/運用鍵/NodeID/セッション再開情報**を `data/` に保存
* **“子端末の安定ID” を UniqueID(0x0012) を主キー**にしてローカルDBに保持（NodeLabelやEndpoint番号が変わっても追従できる）([Matter Survey][2])

### B. 子端末の列挙・同定ロジック（ブリッジ向け）

* Root Endpoint の PartsList で Endpoint を全列挙 ([Matter Survey][1])
* 各 Endpoint で 0x0039 を読めるなら **UniqueID/NodeLabel/Reachable** を取得して一覧化 ([Matter Survey][2])
* 0x0039 が無い Endpoint は “機能Endpoint” の可能性があるので、Descriptor ServerList(0x0001) を読んで搭載Clusterを見て分類（DoorLock / OnOff 等）

### C. 操作系（SESAME 5想定）

* Door Lock Cluster(0x0101) が載っていれば **LockDoor(0x00) / UnlockDoor(0x01)** を送る([CSA-IOT][4])
* 電池は Power Source Cluster(0x002F) の **BatPercentRemaining(0x000C)** を優先して読む（載ってない機種もあるのでフォールバック設計）([Matter Survey][5])

### D. 運用コマンド体系（CLIで完結）

* `setup commission`（初回：QR/手動コード）
* `devices sync`（PartsListから再スキャンしてDB更新）
* `devices list`（子端末一覧表示）
* `devices label set`（NodeLabelを書き込み可能なら名前を固定）
* `lock/unlock`（UniqueID指定で操作）
* `poll` / `watch`（定期取得：電池・到達性など）

---

# go-matter-pack を基準にした「完成までの制作順（推奨マイルストーン）」

※GitHub側の制約で、今この場であなたの fork のコード本文まで正確に読めていません（ディレクトリ構成は確認できています・中身の前提は upstream 構成に寄せて書きます）。そのため **“追加するファイル/責務/関数設計” を具体化**し、あなたの現状コードに合わせる調整は後から差分で可能、という作り方にしています。
（ここからが本題：順番・役割・追加コード・基礎設計）

---

## Milestone 0：プロジェクト方針を固定（1バイナリ・初回だけ設定）

**役割**：後戻りを防ぐ“製品要件”の確定

**追加仕様（確定）**

* 実行モードは2つだけ

  * `setup`：コミッショニング＋初期同期
  * `run`：運用（操作/監視/定期収集）
* データ保存先：`$XDG_STATE_HOME/go-matter-pack/`（無ければ `~/.local/state/...`）
* 外部常駐サービス不要（systemd でこのバイナリだけ常駐させるのはOK）

**成果物**

* `doc/architecture.md`：全体図（後述の設計をコピペでOK）

---

## Milestone 1：内部モジュール分割（“機能を足す場所” を作る）

**役割**：Matter処理とCLI処理、永続化を分離し、以後の追加が楽になる

**追加コード（新規ディレクトリ案）**

* `internal/app/`

  * `config.go`：設定読込（yaml/json/env）
  * `paths.go`：データ保存先解決
* `internal/store/`

  * `store.go`：JSON/bolt/sqlite どれでも良い（最初はJSONでOK）
* `internal/log/`：zap/slogラッパ（任意）

**設計（最低限の型）**

* `type App struct { Ctrl Controller; Store Store; Logger *slog.Logger }`

---

## Milestone 2：Matter “運用接続” 抽象を作る（最重要の土台）

**役割**：下位ライブラリ（go-matterの成熟度）に依存せず、上位機能を先に作る

**追加コード**

* `internal/matterctrl/controller.go`

**基礎設計（インターフェース）**

```go
type Controller interface {
  // ノード（= Hub3）の運用接続ができるか
  Ping(ctx context.Context, nodeID uint64) error

  // 汎用Read/Write/Invoke（ClusterとAttribute/Commandは数値で扱える）
  ReadAttribute(ctx context.Context, nodeID uint64, endpoint uint16, clusterID uint32, attrID uint32) (any, error)
  WriteAttribute(ctx context.Context, nodeID uint64, endpoint uint16, clusterID uint32, attrID uint32, value any) error
  InvokeCommand(ctx context.Context, nodeID uint64, endpoint uint16, clusterID uint32, cmdID uint32, payload any) (any, error)
}
```

**ポイント**

* この時点では **“コミッショニング無しで繋がる前提”** の実装でもOK
* 後で Milestone 6 以降で “コミッショニング（PASE/CASE）” 実装に差し替える

---

## Milestone 3：ブリッジ子端末の列挙（PartsList → 0x0039 読み）

**役割**：あなたが困ってる「Hub3配下のSESAME 5の一覧取得」を実現する核

**追加コード**

* `internal/mattermodel/scan.go`
* `internal/mattermodel/types.go`

**基礎設計（データモデル）**

```go
type BridgedDevice struct {
  NodeID      uint64   // Hub3のNodeID
  Endpoint    uint16   // 子endpoint
  UniqueID    string   // 0x0039 / 0x0012
  NodeLabel   string   // 0x0039 / 0x0005
  Reachable   *bool    // 0x0039 / 0x0011
  Clusters    []uint32 // Descriptor ServerListなどから推定
}
```

**スキャン手順（実装手順そのもの）**

1. `parts := ReadAttribute(node, 0, 0x001D, 0x0003)` で PartsList を取る([Matter Survey][1])
2. 各 `ep in parts` について

   * `uid := ReadAttribute(node, ep, 0x0039, 0x0012)`（UniqueID）([Matter Survey][2])
   * `label := ReadAttribute(node, ep, 0x0039, 0x0005)`（NodeLabel）([Matter Survey][2])
   * `reach := ReadAttribute(node, ep, 0x0039, 0x0011)`（Reachable）([Matter Survey][2])
3. 取れないepは “機能ep” の可能性 → Descriptor の ServerList 等で搭載Clusterを読む（分類用）

**成果物（CLIコマンド）**

* `devices sync --node <nodeid>`：DB更新
* `devices list`：一覧表示（UniqueID/Label/Reachable/Endpoint）

---

## Milestone 4：“困りごと対策”＝安定選択（UniqueID主キー + label固定）

**役割**：Endpoint番号が変わっても、あなたのアプリ側で同じSESAMEを選べる

**追加コード**

* `internal/store/registry.go`

**永続化設計**

* `registry.json` に `unique_id -> {endpoint, last_seen, label, capabilities...}` を保存
* `devices sync` で毎回更新

  * endpoint変化を検知して追従
  * 消えたunique_idは “offline/removed” として保持（履歴が重要）

**任意で強い対策**

* NodeLabel が writable なら、初回にあなたの命名を **Hub3側にも書き込む**

  * `WriteAttribute(node, ep, 0x0039, 0x0005, "Sesame玄関")`([Matter Survey][2])
    → Homeアプリ等とも整合しやすくなる

---

## Milestone 5：操作（Lock/Unlock）を “子端末指定” で通す

**役割**：最小の価値（実際に鍵が動く）を最短で出す

**追加コード**

* `internal/usecase/lock.go`

**設計**

* 入力：`unique_id`（ユーザーはこれだけ指定）
* 内部：`unique_id -> endpoint` を引いて `InvokeCommand`

**Matter操作**

* Door Lock Cluster(0x0101) の `Lock Door(0x00)` / `Unlock Door(0x01)` を送る([CSA-IOT][4])

  * PIN必要な機種に備え、payloadに PINCode を載せられる設計にしておく（未使用なら空でOK）

**CLI**

* `lock --id <unique_id>`
* `unlock --id <unique_id> [--pin 1234]`

---

## Milestone 6：電池・到達性などの定期取得（poll/watch）

**役割**：あなたの「定期的に電池残量や各種データを受信」を実現

**追加コード**

* `internal/metrics/readers/`（プラグイン風に増やせる形）

  * `powersource.go`：0x002F BatPercentRemaining(0x000C) を読む([Matter Survey][5])
  * `reachability.go`：0x0039 Reachable(0x0011) を読む([Matter Survey][2])
* `internal/usecase/poll.go`

**設計**

```go
type MetricReader interface {
  Name() string
  Read(ctx context.Context, ctrl Controller, nodeID uint64, dev BridgedDevice) (map[string]any, error)
}
```

**CLI**

* `poll --interval 10m --out jsonl`（標準出力 or ファイル）
* `poll --once`（動作確認に使う）

---

## Milestone 7：本丸（自前コミッショニング）を入れる／回避策も用意

**役割**：あなたのアプリ単体で “QR/11桁コードから Hub3 をペアリング” を完結させる

ここが一番重いです。理由は go-matter 側の進捗で **PASEが未完成**だから。([GitHub][3])
なので、実装戦略を2本立てにします。

### 7-A（最短ルート）：すでにペアリング済みの情報を取り込む “import”

* 例：別コントローラでコミッショニングだけ済ませ、あなたのアプリは運用接続に必要な情報を取り込む
* ただし、実際に可能かは “どの情報を取り出せるか” に依存（環境差が大きい）

### 7-B（正攻法）：go-matter-pack内にコミッショニング実装を追加

**追加実装が必要になりやすい要素**

* Setup Payload（QR/Base38/Manual code）は go-matter でも実装済み領域がある([GitHub][3])
* その先の **PASE → CASE → NOC/Fabric 保存 → Operational discovery** を実装して永続化する

**コード配置（案）**

* `matter/commissioning/`：PASE/CASE/証明書/ストレージI/F
* `internal/commission/`：CLIから呼ぶユースケース

**CLI**

* `setup commission --qr "MT:...."` or `--manual 34970112332`
* `setup doctor`（commission後に `devices sync` まで自動実行して疎通確認）

---

## Milestone 8：パッケージ運用（systemd + 設定固定）

**役割**：Ubuntuで「導入して放置」を可能にする

**追加物**

* `packaging/systemd/go-matter-pack.service`
* `doc/INSTALL.md`（依存、権限、ファイアウォール注意）

**運用設計**

* `go-matter-pack run --poll-interval 10m --export /var/lib/.../metrics.jsonl`
* ログはjournaldへ

---

# ここまで作れば “あなたの困りごと” はどう解決されるか

* **Hub3配下のSESAME 5一覧**：Descriptor PartsList で列挙できる([Matter Survey][1])
* **どのSESAME 5かの固定**：Bridged Device Basic Info の UniqueID を主キーにDB化できる([Matter Survey][2])
* **開閉信号**：Door Lock の Lock/Unlock コマンドに落とせる([CSA-IOT][4])
* **電池などの定期取得**：Power Source の BatPercentRemaining などを読める（機種差はフォールバックで吸収）([Matter Survey][5])

---
[3]: https://github.com/cybergarage/go-matter?utm_source=chatgpt.com "cybergarage/go-matter"
[4]: https://csa-iot.org/wp-content/uploads/2023/10/Matter-1.2-Application-Cluster-Specification.pdf?utm_source=chatgpt.com "Matter Application Cluster Specification Version 1.2"
[5]: https://www.matter-survey.org/cluster/0x002F "Power Source - Matter Survey"
