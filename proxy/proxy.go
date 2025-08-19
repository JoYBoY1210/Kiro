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
		if strings.HasPrefix(r.URL.Path, "/kiro/route/") {
			tail := strings.TrimPrefix(r.URL.Path, "/kiro/route/")
			parts := strings.SplitN(tail, "/", 2)
			target := parts[0]
			rest := "/"
			if len(parts) == 2 && parts[1] != "" {
				rest = "/" + parts[1]
			}

			if !p.isAllowed(p.ServiceName, target) {
				log.Printf("[BLOCK] %s → %s method=%s path=%s", p.ServiceName, target, r.Method, r.URL.Path)
				http.Error(w, "blocjed by mesh", http.StatusForbidden)
				return
			}
			targetProxyPort, ok := p.ServiceMap[target]
			if !ok {
				log.Printf("[ERROR] service '%s' not found (requested by %s)", target, p.ServiceName)
				http.Error(w, "no such service: "+target, http.StatusNotFound)
				return
			}

			finalUrl := fmt.Sprintf("http://localhost:%d%s", targetProxyPort, rest)
			log.Printf("[ROUTE] %s → %s %s method=%s", p.ServiceName, target, rest, r.Method)

			req, err := http.NewRequest(r.Method, finalUrl, r.Body)
			if err != nil {
				log.Printf("[ERROR] mesh route request creation: %v", err)
				http.Error(w, "proxy Error: "+err.Error(), http.StatusInternalServerError)
				return
			}
			req.Header = r.Header.Clone()
			req.Header.Set("X-Kiro-caller", p.ServiceName)

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				log.Printf("[ERROR] mesh routing: %v", err)
				http.Error(w, "proxy Error: "+err.Error(), http.StatusInternalServerError)
				return
			}
			defer resp.Body.Close()
			for k, v := range resp.Header {
				w.Header()[k] = v
			}
			w.WriteHeader(resp.StatusCode)
			io.Copy(w, resp.Body)
			return
		}

		log.Printf("[PROXY] %s => local:%d %s %s", p.ServiceName, p.TargetPort, r.Method, r.URL.Path)
		req, err := http.NewRequest(r.Method, targetUrl.String()+r.URL.Path, r.Body)
		if err != nil {
			log.Printf("[ERROR] proxy request creation: %v", err)
			http.Error(w, "forwarding error: "+err.Error(), http.StatusBadGateway)
			return
		}
		req.Header = r.Header.Clone()
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Printf("[ERROR] upstream forwarding: %v", err)
			http.Error(w, "forwarding error: "+err.Error(), http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()
		for k, v := range resp.Header {
			w.Header()[k] = v
		}
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
	})

	addr := fmt.Sprint(":", p.ListenPort)
	log.Printf("[START] %s proxy running on %s (target: %d)", p.ServiceName, addr, p.TargetPort)
	http.ListenAndServe(addr, handler)
}

func (p *Proxy) isAllowed(from, to string) bool {
	if m, ok := p.Allowed[from]; ok {
		return m[to]
	}
	return false
}
