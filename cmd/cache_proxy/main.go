package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGTERM, syscall.SIGINT)

	log.Printf("[SYSTEM] Proxy Server Up %s", cfg.RUN_SERVER)

	server := http.Server {
		Addr: cfg.RUN_SERVER,
		Handler: mux,
	}

	go func () {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Fatal system failure down inside HTTP network router: %v", err)
		}
	}()
	<- done

	log.Printf("[SYSTEM] Sutting down the server...")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second * 5)
	defer cancel()

	_err := server.Shutdown(ctx)

	if _err != nil {
		log.Printf("[SYSTEM] Error Occured while closing the server %s", _err.Error())
	}

	log.Printf("[SYSTEM] Server closed")
	
}