package services

import (
	"fmt"
	"io"
	"log"
	"net/http"
)

func StartDashboardService(port int, proxyPort int) {
	mux := http.NewServeMux()

	mux.HandleFunc("/dashboard", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[DASHBOARD] /dashboard called from %s method=%s", r.RemoteAddr, r.Method)
		fmt.Fprintf(w, "hello")
	})

	mux.HandleFunc("/profile", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[DASHBOARD] /profile called from %s method=%s", r.RemoteAddr, r.Method)
		url := fmt.Sprintf("http://localhost:%d/kiro/route/profileService/profile", proxyPort)
		resp, err := http.Get(url)
		if err != nil {
			log.Printf("[DASHBOARD][ERROR] calling profile via proxy failed: %v", err)
			http.Error(w, "calling profile via proxy failed: "+err.Error(), http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()
		w.WriteHeader(resp.StatusCode)
		_, err = io.Copy(w, resp.Body)
		if err != nil {
			log.Printf("[DASHBOARD][ERROR] copying profile response body: %v", err)
		}
	})

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[DASHBOARD] /health called from %s method=%s", r.RemoteAddr, r.Method)
		fmt.Fprintf(w, "ok")
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "blocked by mesh", http.StatusForbidden)
		log.Printf("[BLOCK] %s → %s method=%s path=%s : non-mesh route blocked", "AuthService", "unknown", r.Method, r.URL.Path)
		return
	})

	addr := fmt.Sprintf(":%d", port)
	log.Printf("[DASHBOARD] started on %s with proxy on %d", addr, proxyPort)
	http.ListenAndServe(addr, mux)
}
