package proxy

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)

type Proxy struct {
	ServiceName string
	ListenPort  int
	TargetPort  int
	ServiceMap  map[string]int
	Allowed     map[string]map[string]bool
}

func (p *Proxy) Start() {
	targetUrl, _ := url.Parse(fmt.Sprintf("http://localhost:%d", p.TargetPort))
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host := normalizeHost(r.Host)

		if host == "" || host == p.ServiceName {
			log.Printf("[proxy:%s] LOCAL %s %s -> :%d", p.ServiceName, r.Method, r.URL.Path, p.TargetPort)
			req, _ := http.NewRequest(r.Method, targetUrl.String()+r.URL.Path, r.Body)
			req.Header = r.Header.Clone()
			chain := r.Header.Get("X-Kiro-Caller")
			if chain == "" {
				chain = p.ServiceName
			} else {
				chain = chain + "->" + p.ServiceName
			}
			req.Header.Set("X-Kiro-Caller", chain)
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

		finalUrl := fmt.Sprintf("http://localhost:%d%s", targetProxyPort, r.URL.Path)
		log.Printf("[proxy:%s] ROUTE %s %s -> %s(:%d)", p.ServiceName, r.Method, r.URL.Path, target, targetProxyPort)
		req, err := http.NewRequest(r.Method, finalUrl, r.Body)
		if err != nil {
			http.Error(w, "Internal routing error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		req.Header = r.Header.Clone()
		req.Header.Set("X-Kiro-Caller", p.ServiceName)
		req.Host = target
		resp, err := http.DefaultClient.Do(req)
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
	addr := fmt.Sprintf(":%d", p.ListenPort)
	log.Printf("[proxy:%s] listening on %s -> app :%d", p.ServiceName, addr, p.TargetPort)
	http.ListenAndServe(addr, handler)
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
