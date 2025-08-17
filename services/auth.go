package services

import(
	"fmt"
	"net/http"
)


func StartAuthService(port int){
	mux:=http.NewServeMux()
	mux.HandleFunc("/auth",func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w,"auth servide started")
	})
	mux.HandleFunc("/health",func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w,"ok")
	})
	addr:=fmt.Sprintf(":%d",port)
	fmt.Println("Auth service started on "+addr)
	http.ListenAndServe(addr,mux)
}