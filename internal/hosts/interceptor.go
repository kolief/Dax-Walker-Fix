package hosts

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"daxwalkerfix/internal/idleexit"
	"daxwalkerfix/internal/output"
	"daxwalkerfix/internal/proxy"
	socks "golang.org/x/net/proxy"
)

const (
	domain    = "walker.dax.cloud"
	hostsFile = `C:\Windows\System32\drivers\etc\hosts`
	hostsLine = "127.0.0.1	walker.dax.cloud  # DAX_INTERCEPT"
)

type Interceptor struct {
	proxies []*proxy.Proxy
	debug   bool
	wg      sync.WaitGroup
	counter uint64
}

func New(proxies []*proxy.Proxy, debug bool) *Interceptor {
	return &Interceptor{
		proxies: proxies,
		debug:   debug,
	}
}

func (i *Interceptor) Start(ctx context.Context) error {
	if err := i.addHostsEntry(); err != nil {
		return fmt.Errorf("failed to modify hosts file: %v", err)
	}
	defer i.removeHostsEntry()

	listener, err := net.Listen("tcp", "127.0.0.1:443")
	if err != nil {
		return fmt.Errorf("failed to listen on port 443: %v", err)
	}
	defer listener.Close()

	if i.debug {
		fmt.Printf("Listening on 127.0.0.1:443 for %s traffic\n", domain)
	}

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			i.wg.Add(1)
			go i.handleConnection(conn)
		}
	}()

	<-ctx.Done()
	i.wg.Wait()
	return nil
}

func (i *Interceptor) handleConnection(client net.Conn) {
	defer i.wg.Done()
	defer client.Close()
	
	idleexit.Reset()
	
	var p *proxy.Proxy
	if len(i.proxies) > 0 {
		index := atomic.AddUint64(&i.counter, 1) - 1
		p = i.proxies[index%uint64(len(i.proxies))]
	}
	
	if p != nil {
		proxyType := "SOCKS5"
		if p.Type == proxy.HTTPS {
			proxyType = "HTTPS"
		}
		output.Info("Connection to %s via %s proxy %s", domain, proxyType, p.Address)
	} else {
		output.Info("Connection to %s direct", domain)
	}
	
	if i.debug {
		if p != nil {
			proxyType := "SOCKS5"
			if p.Type == proxy.HTTPS {
				proxyType = "HTTPS"
			}
			fmt.Printf("Connection -> %s via %s proxy\n", domain, proxyType)
		} else {
			fmt.Printf("Connection -> %s direct\n", domain)
		}
	}
	
	target, err := i.connectTo(domain+":443", p)
	if err != nil {
		if i.debug {
			fmt.Printf("Failed to connect: %v\n", err)
		}
		return
	}
	defer target.Close()
	
	go io.Copy(target, client)
	io.Copy(client, target)
}

func (i *Interceptor) connectTo(addr string, p *proxy.Proxy) (net.Conn, error) {
	if p == nil {
		return net.DialTimeout("tcp", addr, 30*time.Second)
	}
	
	switch p.Type {
	case proxy.SOCKS5:
		return i.connectViaSocks5(addr, p)
	case proxy.HTTPS:
		return i.connectViaHTTPS(addr, p)
	default:
		return nil, fmt.Errorf("unsupported proxy type")
	}
}

func (i *Interceptor) connectViaSocks5(addr string, p *proxy.Proxy) (net.Conn, error) {
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
		return nil, err
	}
	
	return dialer.Dial("tcp", addr)
}

func (i *Interceptor) connectViaHTTPS(addr string, p *proxy.Proxy) (net.Conn, error) {
	proxyURL := &url.URL{
		Scheme: "http",
		Host:   p.Address,
	}
	
	if p.Auth != nil {
		proxyURL.User = p.Auth
	}
	
	proxyConn, err := net.DialTimeout("tcp", p.Address, 30*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to proxy: %v", err)
	}
	
	connectReq := &http.Request{
		Method: "CONNECT",
		URL:    &url.URL{Opaque: addr},
		Host:   addr,
		Header: make(http.Header),
	}
	
	if p.Auth != nil {
		if password, ok := p.Auth.Password(); ok {
			connectReq.SetBasicAuth(p.Auth.Username(), password)
		}
	}
	
	if err := connectReq.Write(proxyConn); err != nil {
		proxyConn.Close()
		return nil, fmt.Errorf("failed to send CONNECT request: %v", err)
	}
	
	resp, err := http.ReadResponse(bufio.NewReader(proxyConn), connectReq)
	if err != nil {
		proxyConn.Close()
		return nil, fmt.Errorf("failed to read CONNECT response: %v", err)
	}
	defer resp.Body.Close()
	
    if resp.StatusCode != 200 {
        proxyConn.Close()
        return nil, fmt.Errorf("proxy returned status %d", resp.StatusCode)
    }

    return proxyConn, nil
}

func (i *Interceptor) addHostsEntry() error {
	lines, err := i.readHosts()
	if err != nil {
		return err
	}
	
	for _, line := range lines {
		if strings.Contains(line, domain) && strings.Contains(line, "127.0.0.1") {
			return nil
		}
	}
	
	lines = append(lines, hostsLine)
	return i.writeHosts(lines)
}

func (i *Interceptor) removeHostsEntry() error {
	lines, err := i.readHosts()
	if err != nil {
		return err
	}
	
	var newLines []string
	for _, line := range lines {
		if strings.Contains(line, domain) && strings.Contains(line, "DAX_INTERCEPT") {
			continue
		}
		newLines = append(newLines, line)
	}
	
	return i.writeHosts(newLines)
}

func (i *Interceptor) readHosts() ([]string, error) {
	file, err := os.Open(hostsFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	
	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	
	return lines, scanner.Err()
}

func (i *Interceptor) writeHosts(lines []string) error {
	file, err := os.Create(hostsFile)
	if err != nil {
		return err
	}
	defer file.Close()
	
	writer := bufio.NewWriter(file)
	for _, line := range lines {
		fmt.Fprintln(writer, line)
	}
	
	return writer.Flush()
}