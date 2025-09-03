package bandwidth

import (
	"fmt"
	"sync/atomic"
	"time"

	"daxwalkerfix/internal/output"
)

var (
	bytesIn  int64
	bytesOut int64
	startTime time.Time
)

func Init() {
	startTime = time.Now()
	output.Info("Bandwidth tracking started")
}

func AddIn(bytes int64) {
	atomic.AddInt64(&bytesIn, bytes)
}

func AddOut(bytes int64) {
	atomic.AddInt64(&bytesOut, bytes)
}

func GetStats() (int64, int64, time.Duration) {
	return atomic.LoadInt64(&bytesIn), atomic.LoadInt64(&bytesOut), time.Since(startTime)
}

func FormatBytes(bytes int64) string {
	mb := float64(bytes) / (1024 * 1024)
	if mb < 0.1 {
		return fmt.Sprintf("%.2f MB", mb)
	}
	return fmt.Sprintf("%.1f MB", mb)
}

func FormatBits(bytes int64) string {
	mb := float64(bytes*8) / (1024 * 1024)
	if mb < 0.1 {
		return fmt.Sprintf("%.2f Mb", mb)
	}
	return fmt.Sprintf("%.1f Mb", mb)
}

func LogSession() {
	in, out, duration := GetStats()
	total := in + out
	output.Info("Session ended - Duration: %v, In: %s, Out: %s, Total: %s", 
		duration.Round(time.Second), FormatBytes(in), FormatBytes(out), FormatBytes(total))
}