package proxy

	import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	socks "golang.org/x/net/proxy"
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

func Load() ([]*Proxy, error) {
    home, err := os.UserHomeDir()
    if err != nil {
        return nil, fmt.Errorf("failed to locate user home: %v", err)
    }

    path := filepath.Join(home, "Desktop", "proxy.txt")

    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("failed to read proxy.txt from Desktop: %v", err)
    }
	
	var proxies []*Proxy
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		
        parts := strings.Split(line, ":")
        if len(parts) < 3 {
            continue
        }

        var proxyType ProxyType
        switch strings.ToLower(parts[0]) {
        case "socks5":
            proxyType = SOCKS5
        case "https", "http":
            proxyType = HTTPS
        default:
            continue
        }

        p := &Proxy{
            Address: fmt.Sprintf("%s:%s", parts[1], parts[2]),
            Type:    proxyType,
        }

        if len(parts) >= 5 {
            p.Auth = url.UserPassword(parts[3], strings.Join(parts[4:], ":"))
        }

        if testProxy(p) {
            proxies = append(proxies, p)
        }
	}
	
	if len(proxies) == 0 {
		return nil, fmt.Errorf("no working proxies found")
	}
	
	return proxies, nil
}

func LoadAuto() ([]*Proxy, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to locate user home: %v", err)
	}

	path := filepath.Join(home, "Desktop", "proxy.txt")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read proxy.txt from Desktop: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	untyped := false
	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		lower := strings.ToLower(line)
		if !(strings.HasPrefix(lower, "socks5:") || strings.HasPrefix(lower, "https:") || strings.HasPrefix(lower, "http:")) {
			untyped = true
			break
		}
	}

	if !untyped {
		return Load()
	}

	fmt.Println("Choose the type for your proxies:")
	fmt.Println("1) SOCKS5")
	fmt.Println("2) HTTPS")
	fmt.Print("Press 1 or 2: ")
	choice := readSingleKey()

	defaultType := SOCKS5
	switch strings.TrimSpace(strings.ToLower(choice)) {
	case "2", "https", "http":
		defaultType = HTTPS
	}
	return LoadWithDefault(defaultType)
}

func readSingleKey() string {
	dll := syscall.NewLazyDLL("msvcrt.dll")
	proc := dll.NewProc("_getch")
	r, _, _ := proc.Call()
	ch := byte(r & 0xFF)
	fmt.Println()
	return string(ch)
}

func LoadWithDefault(defaultType ProxyType) ([]*Proxy, error) {
    home, err := os.UserHomeDir()
    if err != nil {
        return nil, fmt.Errorf("failed to locate user home: %v", err)
    }

    path := filepath.Join(home, "Desktop", "proxy.txt")

    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("failed to read proxy.txt from Desktop: %v", err)
    }

    var proxies []*Proxy
    lines := strings.Split(strings.TrimSpace(string(data)), "\n")

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
        var host string
        var port string
        var auth *url.Userinfo

        switch strings.ToLower(parts[0]) {
        case "socks5":
            proxyType = SOCKS5
            if len(parts) < 3 {
                continue
            }
            host, port = parts[1], parts[2]
            if len(parts) >= 5 {
                auth = url.UserPassword(parts[3], strings.Join(parts[4:], ":"))
            }
        case "https", "http":
            proxyType = HTTPS
            if len(parts) < 3 {
                continue
            }
            host, port = parts[1], parts[2]
            if len(parts) >= 5 {
                auth = url.UserPassword(parts[3], strings.Join(parts[4:], ":"))
            }
        default:
            proxyType = defaultType
            host = parts[0]
            port = parts[1]
            if len(parts) >= 4 {
                auth = url.UserPassword(parts[2], strings.Join(parts[3:], ":"))
            }
        }

        p := &Proxy{
            Address: fmt.Sprintf("%s:%s", host, port),
            Auth:    auth,
            Type:    proxyType,
        }

        if testProxy(p) {
            proxies = append(proxies, p)
        }
    }

    if len(proxies) == 0 {
        return nil, fmt.Errorf("no working proxies found")
    }

    return proxies, nil
}

func testProxy(p *Proxy) bool {
	switch p.Type {
	case SOCKS5:
		return testSOCKS5Proxy(p)
	case HTTPS:
		return testHTTPSProxy(p)
	default:
		return false
	}
}

func testSOCKS5Proxy(p *Proxy) bool {
	var auth *socks.Auth
	if p.Auth != nil {
		if password, ok := p.Auth.Password(); ok {
			auth = &socks.Auth{
				User:     p.Auth.Username(),
				Password: password,
			}
		}
	}
	
	dialer, err := socks.SOCKS5("tcp", p.Address, auth, socks.Direct)
	if err != nil {
		return false
	}
	
	conn, err := dialer.Dial("tcp", "httpbin.org:80")
	if err != nil {
		return false
	}
	defer conn.Close()
	
	return true
}

func testHTTPSProxy(p *Proxy) bool {
	proxyURL := &url.URL{
		Scheme: "http",
		Host:   p.Address,
	}
	
	if p.Auth != nil {
		proxyURL.User = p.Auth
	}
	
	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
	}
	
	client := &http.Client{
		Transport: transport,
		Timeout:   10 * time.Second,
	}
	
	resp, err := client.Get("http://httpbin.org/ip")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	
	return resp.StatusCode == 200
}