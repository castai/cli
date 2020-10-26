package config

var GlobalFlags = struct {
	Debug  bool
	ApiURL string
}{
	Debug:  false,
	ApiURL: "https://api.cast.ai/v1",
}
