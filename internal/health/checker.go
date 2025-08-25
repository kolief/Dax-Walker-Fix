package health

import (
	"bufio"
	"context"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"daxwalkerfix/internal/hosts"
	"daxwalkerfix/internal/output"
	"daxwalkerfix/internal/proxy"
)

func CheckProxies(ctx context.Context, initialProxies []*proxy.Proxy, proxyFile string, autoRemove bool, updateCallback func([]*proxy.Proxy), headerCallback func([]*proxy.Proxy, int)) {
	interceptor := hosts.New(nil, false)
	currentProxies := initialProxies

	fullCheckTicker := time.NewTicker(5 * time.Minute)
	randomCheckTicker := time.NewTicker(30 * time.Second)
	retryFailedTicker := time.NewTicker(10 * time.Minute)
	defer fullCheckTicker.Stop()
	defer randomCheckTicker.Stop()
	defer retryFailedTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-fullCheckTicker.C:
			var working []*proxy.Proxy
			var failed []*proxy.Proxy
			for _, p := range currentProxies {
				if testProxy(p, interceptor) {
					working = append(working, p)
				} else {
					fmt.Printf("[%s] Failed proxy: %s\n", time.Now().Format("15:04:05"), p.Address)
					output.Info("Removing failed proxy: %s", p.Address)
					failed = append(failed, p)
				}
			}
			if len(working) != len(currentProxies) {
				logFailedProxies(failed)
				if autoRemove {
					removeProxy(proxyFile, failed)
				}
				updateCallback(working)
				headerCallback(working, len(failed))
				currentProxies = working
			}
		case <-randomCheckTicker.C:
			if len(currentProxies) > 0 {
				randomProxy := currentProxies[rand.Intn(len(currentProxies))]
				if !testProxy(randomProxy, interceptor) {
					fmt.Printf("\n[%s] Random check failed: %s\n", time.Now().Format("15:04:05"), randomProxy.Address)
				}
			}
		case <-retryFailedTicker.C:
			recovered := retryFailedProxies(interceptor)
			if len(recovered) > 0 {
				currentProxies = append(currentProxies, recovered...)
				updateCallback(currentProxies)
				headerCallback(currentProxies, -len(recovered))
			}
		}
	}
}

func retryFailedProxies(interceptor *hosts.Interceptor) []*proxy.Proxy {
	home, _ := os.UserHomeDir()
	failedFile := filepath.Join(home, "Desktop", "DaxWalkerFix", "failed_proxies.txt")
	
	data, err := os.ReadFile(failedFile)
	if err != nil {
		return nil
	}
	
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	var recovered []*proxy.Proxy
	var stillFailed []string
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		
		p := &proxy.Proxy{
			Address: line,
			Type:    proxy.SOCKS5,
		}
		
		if testProxy(p, interceptor) {
			recovered = append(recovered, p)
			output.Info("Recovered proxy: %s", line)
		} else {
			stillFailed = append(stillFailed, line)
		}
	}
	
	if len(recovered) > 0 {
		updateFailedFile(failedFile, stillFailed)
		output.Info("Recovered %d proxies", len(recovered))
	}
	
	return recovered
}

func updateFailedFile(filePath string, failed []string) {
	file, err := os.Create(filePath)
	if err != nil {
		return
	}
	defer file.Close()
	
	for _, line := range failed {
		file.WriteString(line + "\n")
	}
}

func logFailedProxies(failed []*proxy.Proxy) {
	if len(failed) == 0 {
		return
	}
	
	home, _ := os.UserHomeDir()
	daxDir := filepath.Join(home, "Desktop", "DaxWalkerFix")
	os.MkdirAll(daxDir, 0755)
	failedFile := filepath.Join(daxDir, "failed_proxies.txt")
	
	file, err := os.OpenFile(failedFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer file.Close()
	
	for _, p := range failed {
		file.WriteString(p.Address + "\n")
	}
}

func removeProxy(filePath string, failed []*proxy.Proxy) {
	if len(failed) == 0 {
		return
	}
	
	failedAddresses := make(map[string]bool)
	for _, p := range failed {
		failedAddresses[p.Address] = true
	}
	
	file, err := os.Open(filePath)
	if err != nil {
		return
	}
	defer file.Close()
	
	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			lines = append(lines, scanner.Text())
			continue
		}
		
		parts := strings.Split(line, ":")
		if len(parts) >= 2 {
			address := parts[0] + ":" + parts[1]
			if !failedAddresses[address] {
				lines = append(lines, scanner.Text())
			}
		} else {
			lines = append(lines, scanner.Text())
		}
	}
	
	newFile, err := os.Create(filePath)
	if err != nil {
		return
	}
	defer newFile.Close()
	
	for _, line := range lines {
		newFile.WriteString(line + "\n")
	}
}

func testProxy(p *proxy.Proxy, interceptor *hosts.Interceptor) bool {
	conn, err := interceptor.TestProxy("httpbin.org:80", p)
	if err != nil {
		return false
	}
	if conn != nil {
		conn.Close()
	}
	return true
}
