package health

import (
	"context"
	"time"

	"daxwalkerfix/internal/hosts"
	"daxwalkerfix/internal/output"
	"daxwalkerfix/internal/proxy"
)

func CheckProxies(ctx context.Context, initialProxies []*proxy.Proxy, updateCallback func([]*proxy.Proxy)) {
	interceptor := hosts.New(nil, false)
	currentProxies := initialProxies

	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(5 * time.Minute):
			var working []*proxy.Proxy
			for _, p := range currentProxies {
				if testProxy(p, interceptor) {
					working = append(working, p)
				} else {
					output.Info("Removing failed proxy: %s", p.Address)
				}
			}
			if len(working) != len(currentProxies) {
				updateCallback(working)
				currentProxies = working
			}
		}
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
