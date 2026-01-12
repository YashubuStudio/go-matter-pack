# Check-list

## 変更履歴

- プロジェクト方針（実行モード/保存先/運用方針）を `doc/architecture.md` に明文化し、計画リストの1を完了に更新。
- 内部モジュール分割のために `internal/app`、`internal/store`、`internal/log` を追加し、計画リストの2を完了に更新。
- 計画リストの2について機能/全体/構成/デバッグ/進捗の各確認を行い、完了状態を反映。
- Matter運用接続の抽象化として `internal/matterctrl/controller.go` を追加し、計画リストの3を完了に更新。
- ブリッジ子端末の列挙向けに `internal/mattermodel` を追加し、PartsList から UniqueID/NodeLabel/Reachable を取得するスキャン処理を整備したうえで、計画リストの4を完了に更新。
- 計画リストの3と4について機能/全体/構成/デバッグ/進捗の各確認を行い、チェック状態を反映。
- 計画リストの3と4を制作完了状態として反映。
- 安定選択のために UniqueID 主キーの台帳構造とラベル固定用の処理を `internal/store/registry.go` に追加し、計画リストの5を完了に更新。
- Door Lock Cluster の Lock/Unlock 操作用ユースケースを `internal/usecase/lock.go` に追加し、計画リストの6を完了に更新。
