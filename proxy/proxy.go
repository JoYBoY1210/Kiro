package proxy

import (
	"crypto/hmac"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/JoYBoY1210/kiro/security"
)

type Proxy struct {
	ServiceName         string
	ListenPort          int
	TargetPort          int
	ServiceMap          map[string]int
	Allowed             map[string]map[string]bool
	AuthMode            string
	HMACSecret          []byte
	ClockSkew           int
	RequiredLocalCaller bool
	HeaderIdentity      string
	HeaderTimestamp     string
	HeaderSignature     string
	HeaderCallerChain   string
	CertFile            string
	KeyFile             string
	CAFile              string
}

func (p *Proxy) Start() {

	cert, err := tls.LoadX509KeyPair(p.CertFile, p.KeyFile)
	if err != nil {
		log.Fatalf("[proxy:%s] LoadX509KeyPair error: %v", p.ServiceName, err)
		return
	}
	caCert, err := os.ReadFile(p.CAFile)
	if err != nil {
		log.Fatalf("[proxy:%s] Read CA cert error: %v", p.ServiceName, err)
		return
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientCAs:    caCertPool,
		ClientAuth:   tls.RequireAndVerifyClientCert,
		MinVersion:   tls.VersionTLS12,
	}

	targetUrl, _ := url.Parse(fmt.Sprintf("http://%s:%d", p.ServiceName, p.TargetPort))
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		host := normalizeHost(r.Host)

		if host == "" || host == p.ServiceName {

			if strings.EqualFold(p.AuthMode, "hmac") && len(p.HMACSecret) > 0 {
				if err := p.verifyRequest(r); err != nil {
					log.Printf("[proxy:%s] VERIFY FAIL ", p.ServiceName)
					http.Error(w, "unauthorized: "+err.Error(), http.StatusForbidden)
					return
				}
			}

			log.Printf("[proxy:%s] LOCAL %s %s -> :%d", p.ServiceName, r.Method, r.URL.Path, p.TargetPort)
			req, _ := http.NewRequest(r.Method, targetUrl.String()+r.URL.Path, r.Body)
			req.Header = r.Header.Clone()
			chain := r.Header.Get(p.HeaderCallerChain)
			if chain == "" {
				chain = p.ServiceName
			} else {
				chain = chain + "->" + p.ServiceName
			}
			req.Header.Set(p.HeaderCallerChain, chain)
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				http.Error(w, "local forward error: "+err.Error(), http.StatusBadGateway)
			}
			defer resp.Body.Close()
			for k, v := range resp.Header {
				w.Header()[k] = v
			}
			w.WriteHeader(resp.StatusCode)
			io.Copy(w, resp.Body)
			return
		}

		target := host
		if !p.isAllowed(p.ServiceName, target) {
			log.Printf("[proxy:%s] BLOCK %s %s -> %s (policy)", p.ServiceName, r.Method, r.URL.Path, target)
			http.Error(w, "Blcoked by mesh policy", http.StatusForbidden)
			return
		}

		targetProxyPort, ok := p.ServiceMap[target]
		if !ok {
			http.Error(w, "Unknown service: "+target, http.StatusNotFound)
			return
		}

		// if p.RequiredLocalCaller && !isLoopback(r) {
		// 	http.Error(w, "External requests no allowed", http.StatusForbidden)
		// 	return
		// }

		finalUrl := fmt.Sprintf("https://%s:%d%s", target, targetProxyPort, r.URL.Path)

		client := p.mTLSClient()

		log.Printf("[proxy:%s] ROUTE %s %s -> %s(:%d)", p.ServiceName, r.Method, r.URL.Path, target, targetProxyPort)
		req, err := http.NewRequest(r.Method, finalUrl, r.Body)
		if err != nil {
			http.Error(w, "Internal routing error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		req.Header = r.Header.Clone()

		if strings.EqualFold(p.AuthMode, "hmac") && len(p.HMACSecret) > 0 {
			ts := strconv.FormatInt(time.Now().Unix(), 10)
			canonical := security.CanonicalFromRequest(p.ServiceName, target, ts, r)
			sign := security.Sign(p.HMACSecret, canonical)

			req.Header.Set(p.HeaderIdentity, p.ServiceName)
			req.Header.Set(p.HeaderTimestamp, ts)
			req.Header.Set(p.HeaderSignature, sign)

		}

		chain := r.Header.Get(p.HeaderCallerChain)
		if chain == "" {
			chain = p.ServiceName
		} else {
			chain = chain + "->" + p.ServiceName
		}
		req.Header.Set(p.HeaderCallerChain, chain)
		req.Host = target

		resp, err := client.Do(req)
		if err != nil {
			http.Error(w, "Internal routing error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		for k, v := range resp.Header {
			w.Header()[k] = v
		}

		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)

	})
	server := &http.Server{
		Addr:      fmt.Sprintf(":%d", p.ListenPort),
		Handler:   handler,
		TLSConfig: tlsConfig,
	}

	log.Printf("[proxy:%s] listening on :%d -> app :%d", p.ServiceName, p.ListenPort, p.TargetPort)
	log.Fatal(server.ListenAndServeTLS("", ""))
}

func (p *Proxy) mTLSClient() *http.Client {
	cert, err := tls.LoadX509KeyPair(p.CertFile, p.KeyFile)
	if err != nil {
		log.Fatalf("[proxy:%s] failed loading client cert/key: %v", p.ServiceName, err)
	}
	caCert, err := os.ReadFile(p.CAFile)
	if err != nil {
		log.Fatalf("[proxy:%s] failed reading CA cert: %v", p.ServiceName, err)
	}
	caPool := x509.NewCertPool()
	caPool.AppendCertsFromPEM(caCert)

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caPool,
		MinVersion:   tls.VersionTLS12,
	}
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}
}

func (p *Proxy) isAllowed(from, to string) bool {
	if m, ok := p.Allowed[from]; ok {
		return m[to]
	}
	return false
}

func normalizeHost(h string) string {
	h = strings.TrimSpace((strings.ToLower(h)))
	if h == "" {
		return ""
	}
	if i := strings.Index(h, ":"); i > 0 {
		h = h[:i]
	}
	if h == "localhost" || h == "127.0.0.1" {
		return ""
	}
	return h
}

// func isLoopback(r *http.Request) bool {
// 	host, _, err := net.SplitHostPort(r.RemoteAddr)
// 	if err != nil {
// 		return false
// 	}
// 	ip := net.ParseIP(host)
// 	if ip == nil {
// 		return false
// 	}

// 	return ip.IsLoopback()
// }

func (p *Proxy) verifyRequest(r *http.Request) error {
	identity := r.Header.Get(p.HeaderIdentity)
	ts := r.Header.Get(p.HeaderTimestamp)
	sign := r.Header.Get(p.HeaderSignature)

	if identity == "" || ts == "" || sign == "" {
		return fmt.Errorf("Headers missing")
	}

	t, err := strconv.ParseInt(ts, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid timestamp")

	}
	now := time.Now().Unix()
	if (now - t) > int64(p.ClockSkew) {
		return fmt.Errorf("too much time taken")
	}
	canonical := security.CanonicalFromRequest(identity, p.ServiceName, ts, r)
	expected := security.Sign(p.HMACSecret, canonical)
	if !hmac.Equal([]byte(sign), []byte(expected)) {
		return fmt.Errorf("invalid signature")
	}
	return nil
}
