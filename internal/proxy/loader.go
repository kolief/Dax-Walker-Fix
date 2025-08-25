package proxy

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"daxwalkerfix/internal/file"
)

type ProxyType int

const (
	SOCKS5 ProxyType = iota
	HTTPS
)

type Proxy struct {
	Address string
	Auth    *url.Userinfo
	Type    ProxyType
}

func Load() ([]*Proxy, bool, error) {
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
			proxyType = userType
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
