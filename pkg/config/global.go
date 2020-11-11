package config

import (
	"os"
)

var GlobalFlags = struct {
	Debug  bool
	ApiURL string
}{
	Debug:  false,
	ApiURL: "https://api.cast.ai/v1",
}

func ApiURL() string {
	fromEnv := os.Getenv(envApiURL)
	if fromEnv != "" {
		return fromEnv
	}
	return GlobalFlags.ApiURL
}

func Debug() bool {
	fromEnv := os.Getenv(envDebug)
	if fromEnv != "" {
		return fromEnv == "true"
	}
	return GlobalFlags.Debug
}
