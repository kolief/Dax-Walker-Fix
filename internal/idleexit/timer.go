package idleexit

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"daxwalkerfix/internal/output"
)

var (
	lastActivity time.Time
	timeout      time.Duration
	mu           sync.Mutex
	cancel       context.CancelFunc
)

func Start(ctx context.Context, limit time.Duration, cancelFunc context.CancelFunc) {
	mu.Lock()
	lastActivity = time.Now()
	timeout = limit
	cancel = cancelFunc
	mu.Unlock()

	if limit == 0 {
		fmt.Println("Idle timeout: disabled")
		output.Info("Idle timeout: disabled")
	} else {
		fmt.Printf("Idle timeout: %v\n", limit)
		output.Info("Idle timeout: %v", limit)
	}

	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				mu.Lock()
				since := time.Since(lastActivity)
				if timeout > 0 && since >= timeout {
					fmt.Printf("Idle timeout reached, exiting\n")
					output.Info("Idle timeout reached, exiting")
					mu.Unlock()
					cancel()
					os.Exit(0)
				}
				mu.Unlock()
			}
		}
	}()
}

func Reset() {
	mu.Lock()
	lastActivity = time.Now()
	mu.Unlock()
}


