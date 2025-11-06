package NHentai

import (
	"context"
	"net"
	"net/http"
	"time"
)

var httpClient = http.Client{
	Transport: &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: func(dialer *net.Dialer) func(context.Context, string, string) (net.Conn, error) {
			return dialer.DialContext
		}(&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}),
		DisableCompression:    true, // disable gzip
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	},
}
