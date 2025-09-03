package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"daxwalkerfix/internal/bandwidth"
	"daxwalkerfix/internal/file"
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
	bandwidth.Init()
	fmt.Println("Dax Walker Fix by Kolief")
	fmt.Println("Redirects walker.dax.cloud through your proxies")
	fmt.Println()
	output.Info("Started Dax Walker Fix")

	updater.Check()

	fmt.Println("\nLoading proxies...")
	output.Info("Loading proxies")

	proxies, autoRemove, err := proxy.Load()
	if err != nil {
		fmt.Printf("FAILED: %v\n", err)
		fmt.Println("Press Enter to exit")
		output.Error("Failed to load proxies: %v", err)
		fmt.Scanln()
		os.Exit(1)
	}
	
	fmt.Printf("Found %d proxies\n\n", len(proxies))
	fmt.Println("Configuration:")
	fmt.Printf("├─ Proxy file: %s\n", file.GetLastLoadedPath())
	if autoRemove {
		fmt.Println("├─ Auto-remove failed: Yes")
	} else {
		fmt.Println("├─ Auto-remove failed: No")
	}
	fmt.Printf("├─ Idle timeout: %d minutes\n", *timeout)
	fmt.Println("└─ Log file: daxwalkerfix.log")
	
	fmt.Println("\nNetwork Setup:")
	fmt.Println("├─ Listening on: 127.0.0.1:443")
	fmt.Println("├─ Hosts file: Modified")
	fmt.Println("└─ Target domain: walker.dax.cloud → 127.0.0.1")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\nShutting down...")
		bandwidth.LogSession()
		output.Info("Shutdown signal received")
		cancel()
	}()

	idleexit.Start(ctx, time.Duration(*timeout)*time.Minute, cancel)

	interceptor := hosts.New(proxies, false)
	
	var workingProxies = len(proxies)
	var failedProxies = 0
	
	printHeader := func() {
		fmt.Print("\033[H\033[2J")
		fmt.Println("Dax Walker Fix by Kolief")
		fmt.Println("Redirects walker.dax.cloud through your proxies")
		fmt.Println()
		fmt.Printf("Status: Running | Active: %d | Total: %d | Time: %s\n", 
			interceptor.GetConnCount(), interceptor.GetTotalConns(), time.Now().Format("15:04:05"))
		fmt.Printf("Proxies: %d working, %d failed\n", workingProxies, failedProxies)
		
		in, out, duration := bandwidth.GetStats()
		total := in + out
		fmt.Printf("Bandwidth: %s in, %s out, %s total | Session: %v\n", 
			bandwidth.FormatBytes(in), bandwidth.FormatBytes(out), bandwidth.FormatBytes(total), 
			duration.Round(time.Second))
		
		fmt.Println(strings.Repeat("━", 60))
		fmt.Println("Press Ctrl+C to stop")
		fmt.Println()
	}

	updateProxyCount := func(working []*proxy.Proxy, failed int) {
		workingProxies = len(working)
		failedProxies += failed
		printHeader()
	}
	go health.CheckProxies(ctx, proxies, file.GetLastLoadedPath(), autoRemove, interceptor.UpdateProxies, updateProxyCount)

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
	
	fmt.Println("\nProxy Status:")
	for _, p := range proxies {
		proxyType := "SOCKS5"
		if p.Type == proxy.HTTPS {
			proxyType = "HTTPS"
		}
		fmt.Printf("%s (%s) - Ready\n", p.Address, proxyType)
	}
	
	fmt.Printf("\nWorking: %d/%d proxies", len(proxies), len(proxies))
	if autoRemove {
		fmt.Print(" | Auto-removal: Enabled")
	} else {
		fmt.Print(" | Auto-removal: Disabled")
	}
	fmt.Println("\n" + strings.Repeat("━", 70))
	output.Info("Interceptor running on 127.0.0.1:443 looking for %s", "walker.dax.cloud")

	printHeader()
	
	headerTicker := time.NewTicker(5 * time.Second)
	defer headerTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			fmt.Println("\n\nShutting down...")
			bandwidth.LogSession()
			output.Info("Shutting down...")
			return
		case <-headerTicker.C:
			printHeader()
			output.Info("Status: Active connections: %d, Total: %d", interceptor.GetConnCount(), interceptor.GetTotalConns())
		}
	}
}

