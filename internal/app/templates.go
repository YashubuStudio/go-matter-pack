package appctx

type Templates struct {
	MatterctlPath string `json:"matterctl_path"`

	// 引数テンプレ（{placeholder} を埋める）
	ScanArgs     string `json:"scan_args"`      // 例: "scan --format json"
	PairCodeArgs string `json:"pair_code_args"` // 例: "pairing code --node {node} --code {code} --format json"
	PairWiFiArgs string `json:"pair_wifi_args"` // 例: "pairing code-wifi --node {node} --code {code} --ssid {ssid} --pass {pass} --format json"

	// 将来拡張用（read/invokeが見つかったらここに入れていく）
	CustomArgs string `json:"custom_args"` // 任意コマンド欄
}

func DefaultTemplates() Templates {
	return Templates{
		MatterctlPath: `D:\Ollama\projects\go-matter-pack\matterctl.exe`,

		// 確定：scan
		ScanArgs: "scan --format json",

		// pairing subcommands は確定したが、フラグは `pairing code --help` を見て確定したい。
		// なので“最初はプレースホルダのみ”にしてGUIから調整できるようにする。
		PairCodeArgs: "pairing code {node} {code} --format json",
		PairWiFiArgs: "pairing code-wifi {node} {code} {ssid} {pass} --format json",

		CustomArgs: "",
	}
}
