package services

import (
	"fmt"
	"net/http"
	"net/url"
	"time"
)

func MeshClient(myProxyPort int) *http.Client {
	proxyURL, _ := url.Parse(fmt.Sprintf("http://localhost:%d", myProxyPort))
	transport := &http.Transport{
		Proxy:              http.ProxyURL(proxyURL),
		MaxIdleConns:       100,
		IdleConnTimeout:    90 * time.Second,
		DisableCompression: false,
		ForceAttemptHTTP2:  false,
	}
	return &http.Client{
		Transport: transport,
		Timeout:   5 * time.Second,
	}
}
