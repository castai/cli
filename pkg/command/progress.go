package command

import (
	"context"
	"time"

	"github.com/cheggaaa/pb/v3"
)

type ProgressConfig struct {
	Title        string
	TotalTimeETA time.Duration
	TickInterval time.Duration
	StopFunc     func(tick int, bar *pb.ProgressBar) bool
}

func ShowProgress(ctx context.Context, cfg ProgressConfig) {
	tmpl := `{{ bar . "<" "-" (cycle . "↖" "↗" "↘" "↙" ) "." ">"}} {{percent .}} {{string . "eta_title" | green}} {{string . "my_blue_string" | blue}}`
	bar := pb.ProgressBarTemplate(tmpl).Start(int(cfg.TotalTimeETA.Seconds()))
	bar.Set("eta_title", cfg.Title)
	var tick int
	for {
		select {
		case <-time.After(cfg.TickInterval):
			bar.Increment()
			if cfg.StopFunc(tick, bar) {
				bar.Finish()
				return
			}
		case <-ctx.Done():
			return
		}
	}
}
