package main


import (
	"log"
	"net/http"

	proxyHTTP "github.com/drona-gyawali/cache-proxy/pkg/http" 
)

func main() {
	proxyConfig := proxyHTTP.ProxyServerInit(5000)

	mux := http.NewServeMux()

	mux.HandleFunc("/proxy", proxyConfig.ProxyServer)

	log.Println("High-Performance Caching Proxy Booted Securely on :8080")

	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("Fatal system failure down inside HTTP network router: %v", err)
	}
}