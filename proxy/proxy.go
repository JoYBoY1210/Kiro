package proxy

import(
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
)

type Proxy struct{
	ServiceName string
	ListenPort int
	TargetPort int
}

func (p *Proxy)Start(){
	targetUrl,_:=url.Parse(fmt.Sprintf("http://localhost:%d",p.TargetPort))
	handler:=http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s-proxy %s %s",p.ServiceName,r.Method,r.URL.Path)
		proxyReq,_:=http.NewRequest(r.Method,targetUrl.String()+r.URL.Path,r.Body)
		proxyReq.Header=r.Header

		resp,err:=http.DefaultClient.Do(proxyReq)
		if err!=nil{
			http.Error(w,"proxy error: "+err.Error(),http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		for k,v:=range resp.Header{
			w.Header()[k]=v
		}
		w.WriteHeader(resp.StatusCode)

		io.Copy(w,resp.Body)

	})

	addr:=fmt.Sprint(":",p.ListenPort)
	log.Printf("%s-proxy running on %s -> %d", p.ServiceName, addr, p.TargetPort)
	http.ListenAndServe(addr,handler)
}