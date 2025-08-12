package proxy

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	socks "golang.org/x/net/proxy"
)

type Proxy struct {
	Address string
	Auth    *url.Userinfo
}

func Load() ([]*Proxy, error) {
	data, err := os.ReadFile("proxy.txt")
	if err != nil {
		return nil, fmt.Errorf("failed to read proxy.txt: %v", err)
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
		
		p := &Proxy{
			Address: fmt.Sprintf("%s:%s", parts[0], parts[1]),
		}
		
		if len(parts) >= 4 {
			p.Auth = url.UserPassword(parts[2], strings.Join(parts[3:], ":"))
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