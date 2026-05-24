package main

import (
	"log"
	"net/http"
	"os"

	proxyHTTP "github.com/drona-gyawali/cache-proxy/pkg/http"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Printf("Env file unable to load")
		return
	}

	proxyConfig := proxyHTTP.ProxyServerInit(5000, os.Getenv("PROXY_TOKEN"))

	mux := http.NewServeMux()

	mux.HandleFunc("/proxy", proxyConfig.ProxyServer)

	log.Println("High-Performance Caching Proxy Booted Securely on :8080")

	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("Fatal system failure down inside HTTP network router: %v", err)
	}
}