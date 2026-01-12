# 追加仕様（Detailed Spec）

## 0. ゴール定義

### 0.1 ユーザー価値

* Hub3（Matterブリッジ）を1度ペアリングしたら、以後は **Goバイナリ1つ**だけで
  * Hub3配下のSESAME 5一覧を取得・選択
  * Lock/Unlock（開閉）
  * 生死確認（オンライン確認）
  * 定期的な状態／電池情報取得
  * （必要なら）外部へデータ提供（JSON/HTTP/socket）
    ができる。

### 0.2 非ゴール

* 初回コミッショニングを完全GUI化（CLIでOK）
* Homeアプリ等と同等の全機能（部屋割り・自動化連携など）

---

## 1. 方式（重要な設計原則）

### 1.1 Endpoint番号は“識別子として使わない”

ブリッジ配下の子端末（SESAME 5）は **増減や再登録でendpoint番号が変わる可能性がある**ため、

* 永続識別には **stable_id**（後述）を使う
* endpoint番号は **都度再探索してマップを作る**（キャッシュは可だが信用しない）

### 1.2 台帳方式（Registry）

初回登録時に「このSESAME 5は玄関」みたいな **人間が理解できる別名**を付け、以後は stable_id で追跡する。

---

## 2. 機能要件

## 2.1 初期設定（初回のみ）

### A) Fabric/証明書/鍵の生成

* `setup init`
* 生成物を指定ディレクトリに永続化
  例：`~/.config/sesame-matter/`

**成果物**

* Fabric用CA
* Controller証明書／秘密鍵
* 設定ファイル（json/yaml）
* ログディレクトリ

### B) Hub3のコミッショニング（ペアリング）

* 入力：QR または 11桁コード（＋任意でDiscriminator）
* `setup commission --qr <...>` または `--code <11digits>`
* NodeID（device-id）をこちらで固定して割り当て可能にする（例：`0x12344321`）

**成果物**

* Hub3のNodeID、接続に必要な情報（Fabric情報、鍵、等）が保存される
* “再設定不要”が成立する基盤

---

## 2.2 運用コマンド

### A) 生死確認（Probe）

* `hub probe`
* 判定基準：
  1. Hub3へセッション確立できる
  2. 既知の軽いRead（例：Descriptor/Basic/Reachable等）が成功
* 結果：OK/NG と理由（timeout / unreachable / auth fail）

### B) 子端末一覧取得（List）

* `devices list`
* 内部フロー：
  1. Hub3に接続
  2. DescriptorのPartsListを読む → endpoint一覧
  3. 各endpointで「識別情報」を読む → stable_id確定
  4. 台帳と突合して表示名を付与
  5. 未登録は “NEW” として表示

**出力例（CLI）**

* `[OK] 玄関 (stable_id=xxx) endpoint=2 reachable=true`
* `[NEW] ??? (stable_id=yyy) endpoint=5 reachable=true`

### C) 子端末登録（Labeling）

* `devices register --stable-id <id> --name "玄関"`
* 既知 stable_id に別名を付与（台帳に保存）
* 逆に削除：`devices forget --stable-id <id>`

### D) ロック操作（Lock/Unlock）

* `lock --name "玄関"` または `lock --stable-id <id>`
* 内部は「stable_id → endpoint」の解決を毎回行う（または短期キャッシュ＋検証）

### E) 状態取得（Read）

* `devices read --name "玄関" --fields lock,battery`
* 取得対象：
  * ロック状態（LockState）
  * 電池関連（あれば）
  * 到達性（Reachable）

### F) 定期監視（Watch）

* `watch --interval 10m --output json --dest <path|stdout|http>`
* 方式は2段構えが堅い：
  * 基本：Polling（10分ごとにread）
  * 可能なら：Subscribe（イベントが来るものだけ購読）

**監視の必須出力**

* タイムスタンプ
* Hub3到達性
* 各stable_idの到達性
* lock状態
* battery（取得できる場合）
* 失敗理由（あれば）

---

## 2.3 外部連携（パッケージ導入用）

あなたの「導入した際にこれらを渡す用の物」にするなら、選択肢は2つ。

### Option 1: ファイル出力（最も簡単・堅牢）

* `watch --output jsonl --dest /var/lib/sesame-matter/state.jsonl`
* 他プロセスが読む（tail/パーサ）

### Option 2: ローカルHTTP（必要なら）

* `serve --listen 127.0.0.1:8787`
* `GET /health`
* `GET /devices`
* `POST /lock`, `POST /unlock`

> “常時サーバが嫌”なら Option 1 が最小。
> ただし `watch` は常駐1プロセスなので、実運用の面倒さはほぼ増えません。

---

## 3. データ設計（永続化）

## 3.1 ディレクトリ構成（例）

`~/.config/sesame-matter/`

* `config.json`（共通設定）
* `fabric/`（証明書・鍵・Fabric情報）
* `registry.json`（台帳：stable_id→表示名など）
* `cache.json`（最後に見た endpoint など短期キャッシュ）
* `logs/`

## 3.2 台帳（registry.json）案

```json
{
  "hub": {
    "node_id": "0x12344321",
    "name": "SesameHub3"
  },
  "devices": {
    "stable_id_ABC": {
      "name": "玄関",
      "last_seen_endpoint": 2,
      "last_seen_at": "2026-01-12T10:00:00Z"
    },
    "stable_id_DEF": {
      "name": "勝手口",
      "last_seen_endpoint": 5,
      "last_seen_at": "2026-01-12T10:00:00Z"
    }
  }
}
```

## 3.3 stable_id の優先順位仕様

1. Bridged Device Basic Information の UniqueID が取れるならそれ
2. 取れない場合：NodeLabel＋Vendor/Productなどの組み合わせ
3. 最悪：endpoint＋特徴量（ただし再登録で崩れるので「仮ID扱い」）

---

## 4. 例外系・堅牢化仕様

* Hub3が落ちている：watchはリトライ＋指数バックオフ（最大間隔上限あり）
* 子端末が減った：台帳から消さず「見えない状態」として記録（復活時に再リンク）
* stable_idが取れない：そのendpointは “UNIDENTIFIED” として表示し、登録不可扱い（または仮登録）
* 操作対象が複数一致（名前重複など）：エラーで止め、候補一覧を提示
* ロック操作の安全：
  * 直前に reachability を確認
  * コマンド失敗時の理由を保存

---

# 制作工程（Implementation Roadmap）

ここからは「作る順番」を、失敗しにくい順に並べます。
各フェーズは **動く成果物**が必ず出るようにします。

---

## フェーズ0：開発環境の固定（0.5〜1日）

**成果物**

* Ubuntu上でGoビルドできる
* 依存（Matterライブラリ）がビルドできる
* `make test` / `go test` が通る

**作業**

* Goバージョン固定（例：1.22+）
* リポジトリ雛形
* ログ基盤（zap/zerologなど）
* 設定読み書き（viperでも自前でも）

---

## フェーズ1：初期設定の骨組み（1〜2日）

**成果物**

* `sesame-matter setup init` が実行でき、設定ディレクトリが作られる
* 鍵・証明書・Fabric情報の生成（最低限）
* `config.json` が出る

**実装ポイント**

* ファイル権限（秘密鍵は600）
* “再設定しない”ために生成物は壊さない（上書きは `--force` のときだけ）

---

## フェーズ2：コミッショニング（ペアリング）機能（1〜3日）

**成果物**

* `setup commission --qr/--code` が成功し、以後再設定不要になる
* Hub3のNodeIDが保存される

**確認**

* コミッショニング後、次回起動で “Hub3に再接続”できる

**落とし穴対策**

* 「最初だけCLI自由」の条件なので、ここは多少手間でもOK
* ここが通れば後が全部ラク

---

## フェーズ3：Hub生死確認（probe）（0.5〜1日）

**成果物**

* `hub probe` でOK/NGが分かる
* NG理由がログ＆CLIに出る

**実装**

* 接続→軽いRead→終了
* タイムアウト設定（短すぎると不安定、長すぎるとストレス）

---

## フェーズ4：ブリッジ配下一覧（PartsList→endpoint列挙）（1〜2日）

**成果物**

* `devices list` が endpoint一覧を表示できる（この時点はまだ “名前なし” でもOK）

**実装**

* Descriptor cluster の PartsList read
* endpointを順に走査
* “ロックっぽいendpoint”を検出（DoorLock clusterの有無など）

---

## フェーズ5：stable_id 取得＆台帳（Registry）導入（1〜3日）

**成果物**

* `devices list` が stable_id を出せる
* `devices register` で “玄関”などの名前付けができる
* 再起動しても保持される
* endpointが変わっても stable_id で追跡できる（ここで“番号地獄”解消）

**実装**

* Bridged Device Basic Information から UniqueID/NodeLabel等を読む
* `registry.json` 更新
* `cache.json` に last_seen_endpoint を入れる（便利）

---

## フェーズ6：ロック操作（Lock/Unlock）実装（1〜2日）

**成果物**

* `lock --name 玄関` が動く
* `unlock --stable-id ...` が動く
* 失敗時の理由が分かる

**実装**

* 直前に endpoint解決（list相当）
* DoorLock cluster invoke
* 結果確認（必要ならLockState read）

---

## フェーズ7：定期監視（Polling）と出力（1〜3日）

**成果物**

* `watch --interval 10m --dest ...` が動作し続ける
* JSON/JSONLで状態が出る
* systemdで自動起動できる（任意）

**実装**

* tickごとに
  * hub probe
  * devices list（endpoint再解決）
  * 各デバイス read（lock state / battery / reachable）
* 出力先：stdout / ファイル
* backoff：hub不在時は間隔を伸ばす

---

## フェーズ8：Subscribe（イベント受信）で改善（任意・2〜5日）

**成果物**

* ロック状態変化などがリアルタイムに近く取れる（対応している場合）
* Pollingと併用して欠落に強くする

**注意**

* デバイス側がイベントを出してない場合もあるので、まずPollingで完成させてからでOK

---

## フェーズ9：パッケージ化（1〜2日）

**成果物**

* 1バイナリ配布（Go build）
* `deb` か `tar.gz` を作成
* systemd unit例（任意）

**作業**

* `--data-dir` 指定対応（`/var/lib/...` に置けるように）
* 初回セットアップの手順書（README）

---

# 受け入れ基準（Acceptance Criteria）

最低限これが満たせたら「目的達成」です。

1. 初回 `setup init` → `setup commission` が成功
2. 再起動後も再設定不要で `hub probe` が通る
3. `devices list` がブリッジ配下の複数SESAME 5を列挙できる
4. `devices register` で名前付け → 次回も維持
5. `lock/unlock` が名前指定で動く
6. `watch` が定期的に状態（＋可能なら電池）を出力し続ける
7. Hubや子端末が落ちても落ち方が分かり、復帰したら自動復帰する

---

# 最初に作るべき“最小プロトタイプ”（MVP）

最速で価値が出る最小セットはこれです：

* `setup init`
* `setup commission`
* `devices list`（stable_idが取れるところまで）
* `devices register`
* `lock/unlock`
* `hub probe`

このMVPができた時点で、あなたの「番号変更で困る」はほぼ解消します。
その後に `watch`（監視）を足すのが順序として安全です。
