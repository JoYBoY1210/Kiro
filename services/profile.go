package services

import (
	"fmt"
	"log"
	"net/http"
)

func StartProfileService(port int, name string, proxyPort int, certFile, keyFile, caFile string) {
	mux := http.NewServeMux()

	mux.HandleFunc("/profile", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[PROFILE] /profile called from %s method=%s", r.RemoteAddr, r.Method)

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"Service":"Profile","message":"profile data here"}`)
	})

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[PROFILE] /health called from %s method=%s", r.RemoteAddr, r.Method)
		fmt.Fprintf(w, "ok")
	})

	addr := fmt.Sprintf("127.0.0.1:%d", port)
	log.Printf("[PROFILE] Profile service started on %s with proxy on %d", addr, proxyPort)
	http.ListenAndServe(addr, mux)
}
