package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"daxwalkerfix/internal/health"
	"daxwalkerfix/internal/hosts"
	"daxwalkerfix/internal/idleexit"
	"daxwalkerfix/internal/output"
	"daxwalkerfix/internal/proxy"
	"daxwalkerfix/internal/updater"
)

func main() {
	timeout := flag.Int("timeout", 30, "Shutdown after N minutes of inactivity")
	flag.Parse()

	output.InitLogger()
	fmt.Println("Dax Walker Fix by Kolief")
	fmt.Println("Redirects walker.dax.cloud through your proxies")
	fmt.Println()
	output.Info("Started Dax Walker Fix")

	updater.Check()

	fmt.Println("Loading proxies...")
	output.Info("Loading proxies")

	proxies, err := proxy.Load()
	if err != nil {
		fmt.Printf("FAILED: %v\n", err)
		fmt.Println("Press Enter to exit")
		output.Error("Failed to load proxies: %v", err)
		fmt.Scanln()
		os.Exit(1)
	}
	fmt.Printf("Found %d proxies\n", len(proxies))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\nShutting down...")
		output.Info("Shutdown signal received")
		cancel()
	}()

	idleexit.Start(ctx, time.Duration(*timeout)*time.Minute, cancel)

	interceptor := hosts.New(proxies, false)
	go health.CheckProxies(ctx, proxies, interceptor.UpdateProxies)

	go func() {
		err := interceptor.Start(ctx)
		if err != nil {
			fmt.Printf("FAILED: %v\n", err)
			fmt.Println("Try running as Administrator")
			fmt.Println("Press Enter to exit")
			output.Error("Failed to start: %v", err)
			fmt.Scanln()
			os.Exit(1)
		}
	}()

	time.Sleep(500 * time.Millisecond)
	socksCount := 0
	httpsCount := 0
	for _, p := range proxies {
		if p.Type == proxy.SOCKS5 {
			socksCount++
		} else if p.Type == proxy.HTTPS {
			httpsCount++
		}
	}
	if socksCount > 0 && httpsCount == 0 {
		fmt.Printf("Using %d SOCKS5 proxies\n", socksCount)
	} else if httpsCount > 0 && socksCount == 0 {
		fmt.Printf("Using %d HTTPS proxies\n", httpsCount)
	} else {
		fmt.Printf("Using %d proxies (SOCKS5: %d, HTTPS: %d)\n", len(proxies), socksCount, httpsCount)
	}
	fmt.Println("\nStatus: Running (Press Ctrl+C to stop)")
	output.Info("Interceptor running on 127.0.0.1:443 looking for %s", "walker.dax.cloud")

	statusTicker := time.NewTicker(30 * time.Second)
	defer statusTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			fmt.Println("\nShutting down...")
			output.Info("Shutting down...")
			return
		case <-statusTicker.C:
			timeStr := time.Now().Format("15:04:05")
			output.Info("Status: Active - %s", timeStr)
		}
	}
}

