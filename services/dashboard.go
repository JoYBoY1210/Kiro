package services

import(
	"fmt"
	"net/http"
)


func StartDashboardService(port int){
	mux:=http.NewServeMux()
	mux.HandleFunc("/dashboard",func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w,"hello")
	})
	mux.HandleFunc("/health",func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w,"ok")
	})
	addr:=fmt.Sprintf(":%d",port)
	fmt.Println("Dashboard service started on "+addr)
	http.ListenAndServe(addr,mux)
}