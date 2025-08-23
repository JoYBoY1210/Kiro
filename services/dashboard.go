package services

import (
	"fmt"
	"io"
	"log"
	"net/http"
)

func StartDashboardService(port int, proxyPort int) {
	mux := http.NewServeMux()
	client := MeshClient(proxyPort)

	mux.HandleFunc("/dashboard", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[DASHBOARD] /dashboard called from %s method=%s", r.RemoteAddr, r.Method)
		fmt.Fprintf(w, "hello")
	})

	mux.HandleFunc("/profile", func(w http.ResponseWriter, r *http.Request) {
		resp, err := client.Get("http://profileService/profile")
		if err != nil {
			http.Error(w, "dashboard->profile failed: "+err.Error(), http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
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
