package proxy

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"daxwalkerfix/internal/file"
	"daxwalkerfix/internal/output"
)

type ProxyType int

const (
	SOCKS5 ProxyType = iota
	HTTPS
	
	httpTimeout = 30 * time.Second
	maxProxies  = 200
)

type Proxy struct {
	Address string
	Auth    *url.Userinfo
	Type    ProxyType
}

var proxySources = []string{
	"https://raw.githubusercontent.com/TheSpeedX/PROXY-List/master/socks5.txt",
	"https://raw.githubusercontent.com/TheSpeedX/PROXY-List/master/http.txt",
	"https://raw.githubusercontent.com/hookzof/socks5_list/master/proxy.txt",
	"https://raw.githubusercontent.com/clarketm/proxy-list/master/proxy-list-raw.txt",
	"https://api.proxyscrape.com/v2/?request=get&protocol=socks5&timeout=10000&country=all&ssl=all&anonymity=all",
	"https://api.proxyscrape.com/v2/?request=get&protocol=http&timeout=10000&country=all&ssl=all&anonymity=all",
}

func Load() ([]*Proxy, bool, error) {
	fmt.Println("Proxy source: 1) File  2) Scraped Proxies")
	fmt.Print("Enter 1 or 2: ")
	var sourceChoice string
	fmt.Scanln(&sourceChoice)
	
	if sourceChoice == "2" {
		return loadFromInternet()
	}
	
	var path string
	var rememberedType int
	
	rememberedPath, rememberedType := file.LoadPathWithType()
	if rememberedPath != "" {
		path = rememberedPath
	} else {
		fmt.Println("Select proxy.txt file...")
		selectedPath, err := file.SelectProxyFile()
		if err != nil {
			return nil, false, fmt.Errorf("failed to select proxy file: %v", err)
		}
		path = selectedPath
		rememberedType = -1
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, false, fmt.Errorf("failed to read proxy file: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	needsChoice := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if !strings.HasPrefix(strings.ToLower(line), "socks5:") && !strings.HasPrefix(strings.ToLower(line), "https:") {
			needsChoice = true
			break
		}
	}

	var userType ProxyType
	if needsChoice {
		if rememberedType != -1 {
			userType = ProxyType(rememberedType)
		} else {
			fmt.Println("Choose proxy type: 1) SOCKS5  2) HTTPS")
			fmt.Print("Enter 1 or 2: ")
			var choice string
			fmt.Scanln(&choice)
			if choice == "2" {
				userType = HTTPS
			} else {
				userType = SOCKS5
			}
		}
	}

	proxies := parseProxies(lines, userType)

	if len(proxies) == 0 {
		return nil, false, fmt.Errorf("no proxies found")
	}

	fmt.Println("Auto-remove failed proxies? 1) Yes  2) No")
	fmt.Print("Enter 1 or 2: ")
	var removeChoice string
	fmt.Scanln(&removeChoice)
	autoRemove := removeChoice == "1"

	file.SavePathWithType(path, int(userType))
	file.SetLastLoadedPath(path)
	return proxies, autoRemove, nil
}

func parseProxies(lines []string, defaultType ProxyType) []*Proxy {
	var proxies []*Proxy
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.Split(line, ":")
		if len(parts) < 2 {
			continue
		}

		var proxyType ProxyType
		var host, port string
		var auth *url.Userinfo

		if strings.ToLower(parts[0]) == "socks5" {
			if len(parts) < 3 {
				continue
			}
			proxyType = SOCKS5
			host, port = parts[1], parts[2]
			if len(parts) >= 5 {
				auth = url.UserPassword(parts[3], parts[4])
			}
		} else if strings.ToLower(parts[0]) == "https" {
			if len(parts) < 3 {
				continue
			}
			proxyType = HTTPS
			host, port = parts[1], parts[2]
			if len(parts) >= 5 {
				auth = url.UserPassword(parts[3], parts[4])
			}
		} else {
			proxyType = defaultType
			host, port = parts[0], parts[1]
			if len(parts) >= 4 {
				auth = url.UserPassword(parts[2], parts[3])
			}
		}

		p := &Proxy{
			Address: fmt.Sprintf("%s:%s", host, port),
			Auth:    auth,
			Type:    proxyType,
		}
		proxies = append(proxies, p)
	}
	return proxies
}

func loadFromInternet() ([]*Proxy, bool, error) {
	fmt.Println("Downloading proxies...")
	output.Info("Starting proxy download from sources")
	
	client := &http.Client{Timeout: httpTimeout}
	var allLines []string
	
	for _, url := range proxySources {
		fmt.Printf("Fetching from %s...\n", url)
		resp, err := client.Get(url)
		if err != nil {
			fmt.Printf("Failed to fetch from %s: %v\n", url, err)
			output.Warn("Failed to fetch proxies from %s: %v", url, err)
			continue
		}
		
		data, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			fmt.Printf("Failed to read from %s: %v\n", url, err)
			output.Warn("Failed to read proxies from %s: %v", url, err)
			continue
		}
		
		lines := strings.Split(strings.TrimSpace(string(data)), "\n")
		allLines = append(allLines, lines...)
		fmt.Printf("Got %d lines from %s\n", len(lines), url)
		
		if len(allLines) >= maxProxies {
			allLines = allLines[:maxProxies]
			break
		}
	}
	
	proxies := parseProxies(allLines, SOCKS5)
	
	if len(proxies) == 0 {
		return nil, false, fmt.Errorf("no proxies downloaded")
	}
	
	fmt.Printf("Downloaded %d proxies\n", len(proxies))
	output.Info("Downloaded %d proxies from internet sources", len(proxies))
	
	fmt.Println("Auto-remove failed proxies? 1) Yes  2) No")
	fmt.Print("Enter 1 or 2: ")
	var removeChoice string
	fmt.Scanln(&removeChoice)
	autoRemove := removeChoice == "1"
	
	return proxies, autoRemove, nil
}
