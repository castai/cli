package command

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/cheggaaa/pb/v3"
)

func TestProgress(t *testing.T) {
	stop := time.After(3 * time.Second)
	cfg := ProgressConfig{
		Title:        fmt.Sprintf("Estimated time: %s", 5*time.Minute),
		TotalTimeETA: 5 * time.Second,
		TickInterval: 1 * time.Second,
		StopFunc: func(tick int, bar *pb.ProgressBar) bool {
			select {
			case <-stop:
				return true
			default:
				return false
			}
		},
	}

	ctx := context.Background()
	ShowProgress(ctx, cfg)
}
