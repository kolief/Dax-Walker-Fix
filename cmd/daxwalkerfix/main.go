package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"daxwalkerfix/internal/hosts"
	"daxwalkerfix/internal/idleexit"
	"daxwalkerfix/internal/output"
	"daxwalkerfix/internal/proxy"
)

func main() {
	timeout := flag.Int("timeout", 5, "Shutdown after N minutes of inactivity")
	debug := flag.Bool("debug", false, "Enable debug logging")
	flag.Parse()

	output.InitLogger()
	fmt.Println("Dax Walker Fix by Kolief - Hosts file interceptor")
	fmt.Println("Redirects walker.dax.cloud through SOCKS5 proxies")
	fmt.Println()
	output.Info("Dax Walker Fix by Kolief - Hosts file interceptor")
	output.Info("Redirects walker.dax.cloud through SOCKS5 proxies")

	fmt.Print("Loading proxies...")
	output.Info("Loading proxies...")
	proxies, err := proxy.Load()
	if err != nil {
		fmt.Printf("FAILED: %v\n", err)
		fmt.Println("\nMake sure proxy.txt exists with working SOCKS5 proxies.")
		fmt.Println("Press Enter to exit...")
		output.Error("Failed to load proxies: %v", err)
		output.Info("Make sure proxy.txt exists with working SOCKS5 proxies")
		fmt.Scanln()
		os.Exit(1)
	}
	fmt.Printf(" found %d working proxies\n", len(proxies))
	output.Info("Found %d working proxies", len(proxies))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\nShutdown signal received")
		output.Info("Shutdown signal received")
		cancel()
	}()

	idleexit.Start(ctx, time.Duration(*timeout)*time.Minute, cancel)

	fmt.Printf("Starting interceptor (timeout: %d minutes)...\n", *timeout)
	output.Info("Starting interceptor (timeout: %d minutes)", *timeout)
	interceptor := hosts.New(proxies, *debug)
	
	go func() {
		err := interceptor.Start(ctx)
		if err != nil {
			fmt.Printf("FAILED TO START: %v\n", err)
			fmt.Println("\nCommon causes:")
			fmt.Println("1. Not running as Administrator")
			fmt.Println("2. Port 443 already in use")
			fmt.Println("3. Hosts file is read-only")
			fmt.Println("\nPress Enter to exit...")
			output.Error("Failed to start: %v", err)
			output.Info("Common causes: 1. Not running as Administrator, 2. Port 443 already in use, 3. Hosts file is read-only")
			fmt.Scanln()
			os.Exit(1)
		}
	}()

	time.Sleep(500 * time.Millisecond)
	fmt.Println("✓ Interceptor running")
	fmt.Printf("✓ Listening on 127.0.0.1:443 for %s\n", "walker.dax.cloud")
	fmt.Printf("✓ Using %d SOCKS5 proxies\n", len(proxies))
	fmt.Println("\nStatus: Running (Press Ctrl+C to stop)")
	output.Info("Interceptor running")
	output.Info("Listening on 127.0.0.1:443 for %s", "walker.dax.cloud")
	output.Info("Using %d SOCKS5 proxies", len(proxies))
	output.Info("Status: Running")
	
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