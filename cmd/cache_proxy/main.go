package main

import (
	"log"
	"net/http"
	"os"

	"github.com/drona-gyawali/cache-proxy/pkg/config"
	proxyHTTP "github.com/drona-gyawali/cache-proxy/pkg/http"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Printf("Env file unable to load")
		return
	}

	cfg := config.MustLoad()
	proxyConfig := proxyHTTP.ProxyServerInit(cfg.CAPACITY, os.Getenv("PROXY_TOKEN"), cfg.ALLOWED_CLUSTERS)

	mux := http.NewServeMux()

	mux.HandleFunc("/proxy", proxyConfig.ProxyServer)

	log.Printf("Cluster Booted %s", cfg.RUN_SERVER)

	if err := http.ListenAndServe(cfg.RUN_SERVER, mux); err != nil {
		log.Fatalf("Fatal system failure down inside HTTP network router: %v", err)
	}
}