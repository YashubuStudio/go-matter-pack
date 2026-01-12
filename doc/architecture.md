# アーキテクチャ概要

## 実行モード

本プロジェクトの実行モードは 2 つに固定します。

* `setup`: 初回コミッショニングと初期同期のみを実行
* `run`: 運用モード（操作・監視・定期収集）

## データ保存先

データは常に XDG Base Directory を基準に保存します。

* `XDG_STATE_HOME` がある場合: `$XDG_STATE_HOME/go-matter-pack/`
* ない場合: `~/.local/state/go-matter-pack/`

保存対象は Fabric 情報、証明書・鍵、NodeID、セッション再開情報などの運用データです。

## 運用方針

* 外部常駐サービスは不要（1 バイナリで完結）
* systemd で本バイナリを常駐させる構成は許容
* 運用中の CLI 操作はすべて `run` モードに統一
