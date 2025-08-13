package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"daxwalkerfix/internal/hosts"
	"daxwalkerfix/internal/idleexit"
	"daxwalkerfix/internal/output"
	"daxwalkerfix/internal/proxy"
)

func main() {
	timeout := flag.Int("timeout", 30, "Shutdown after N minutes of inactivity")
	debug := flag.Bool("debug", false, "Enable debug logging")
	flag.Parse()

	output.InitLogger()
	fmt.Println("Dax Walker Fix by Kolief - Hosts file interceptor")
	fmt.Println("Redirects walker.dax.cloud through proxies")
	fmt.Println()
	output.Info("Dax Walker Fix by Kolief - Hosts file interceptor")
	output.Info("Redirects walker.dax.cloud through proxies")

    ensureProxyTemplate()

	fmt.Println("Loading proxies...")
	output.Info("Loading proxies...")

    proxies, err := proxy.LoadAuto()
	if err != nil {
		fmt.Printf("FAILED: %v\n", err)
		fmt.Println("\nMake sure proxy.txt exists on your Desktop with working proxies.")
		fmt.Println("Press Enter to exit...")
		output.Error("Failed to load proxies: %v", err)
		output.Info("Make sure proxy.txt exists on Desktop with working proxies")
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
		fmt.Printf("✓ Using %d SOCKS5 proxies\n", socksCount)
		output.Info("Using %d SOCKS5 proxies", socksCount)
	} else if httpsCount > 0 && socksCount == 0 {
		fmt.Printf("✓ Using %d HTTPS proxies\n", httpsCount)
		output.Info("Using %d HTTPS proxies", httpsCount)
	} else {
		fmt.Printf("✓ Using %d proxies (SOCKS5: %d, HTTPS: %d)\n", len(proxies), socksCount, httpsCount)
		output.Info("Using %d proxies (SOCKS5: %d, HTTPS: %d)", len(proxies), socksCount, httpsCount)
	}
	fmt.Println("\nStatus: Running (Press Ctrl+C to stop)")
	output.Info("Interceptor running")
	output.Info("Listening on 127.0.0.1:443 for %s", "walker.dax.cloud")
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

const ProxyTemplate = `# SOCKS5 OR HTTPS proxy list (one per line)
#
# Formats supported:
#   host:port
#   username:password@host:port
#
# Notes:
# - Only SOCKS5 or HTTPS proxies are supported.
# - Lines starting with # are comments and ignored.
# - Blank lines are ignored.
# - The app validates each proxy with a short TCP check before use.
#
# Examples:
# 127.0.0.1:1080
# 203.0.113.10:1080
# 192.168.1.100:1080:username:password

# Add your SOCKS5 or HTTPS proxies here
`

func ensureProxyTemplate() {
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}
	desktop := filepath.Join(home, "Desktop")
	dest := filepath.Join(desktop, "proxy.txt")

	if _, err := os.Stat(dest); err == nil {
		return
	}

	if err := os.WriteFile(dest, []byte(ProxyTemplate), 0644); err == nil {
		fmt.Println("Proxy.txt was created and can be located here: %USERPROFILE%\\Desktop\\proxy.txt")
		output.Info("Proxy.txt was created and can be located here: %%USERPROFILE%%\\Desktop\\proxy.txt")
	}
}