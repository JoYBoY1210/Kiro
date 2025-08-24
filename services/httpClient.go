package services

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"
)

func MeshClient(myProxyPort int, name, certFile, keyFile, caFile string) *http.Client {
	proxyURL, _ := url.Parse(fmt.Sprintf("https://%s:%d", name, myProxyPort))
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		panic(fmt.Sprintf("Failed to load client cert/key: %v", err))
	}
	caCert, err := os.ReadFile(caFile)
	if err != nil {
		panic(fmt.Sprintf("Failed to read CA cert: %v", err))
	}
	caPool := x509.NewCertPool()
	caPool.AppendCertsFromPEM(caCert)

	tlsConfig := &tls.Config{
		Certificates:       []tls.Certificate{cert},
		RootCAs:            caPool,
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: false,
	}

	transport := &http.Transport{
		Proxy:              http.ProxyURL(proxyURL),
		TLSClientConfig:    tlsConfig,
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
