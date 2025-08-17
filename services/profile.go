package services

import (
	"fmt"
	"net/http"
)

func StartProfileService(port int) {
	mux := http.NewServeMux()
	mux.HandleFunc("/profile", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"Service:"Profile","message":"profile data here"}`)
	})
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "ok")
	})
	addr := fmt.Sprintf(":%d", port)
	fmt.Println("Profile service started on " + addr)
	http.ListenAndServe(addr, mux)
}
