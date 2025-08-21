package services

import (
	"fmt"
	"io"
	"log"
	"net/http"
)

func StartAuthService(port int, proxyPort int) {
	mux := http.NewServeMux()
	client := MeshClient(proxyPort)

	mux.HandleFunc("/auth", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[AUTH] /auth endpoint called from %s method=%s", r.RemoteAddr, r.Method)
		fmt.Fprintf(w, "auth service started")
	})

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[AUTH] /health endpoint called from %s method=%s", r.RemoteAddr, r.Method)
		fmt.Fprintf(w, "ok")
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "blocked by mesh", http.StatusForbidden)
		log.Printf("[BLOCK] %s → %s method=%s path=%s : non-mesh route blocked", "AuthService", "unknown", r.Method, r.URL.Path)
		return
	})

	mux.HandleFunc("/dashboard", func(w http.ResponseWriter, r *http.Request) {
		resp, err := client.Get("http://profileService/profile")
		if err != nil {
			http.Error(w, "auth->dashboard failed: "+err.Error(), http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
	})

	addr := fmt.Sprintf(":%d", port)
	log.Printf("[AUTH] Auth service started on %s with proxy on %d", addr, proxyPort)
	http.ListenAndServe(addr, mux)
}
