package main

import (
	"context"
	"flag"
	"fmt"
	"io"
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

	port := flag.Int("port", 8080, "The Proxy Cluster that run on the port.")
	origin := flag.String("origin", "", "Target url you want to cahce")
	flag.Parse()

	cfg := config.MustLoad()
	proxyConfig := proxyHTTP.ProxyServerInit(cfg.CAPACITY, os.Getenv("PROXY_TOKEN"), cfg.ALLOWED_CLUSTERS)
	mux := http.NewServeMux()
	mux.HandleFunc("/proxy", proxyConfig.ProxyServer)

	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGTERM, syscall.SIGINT)

	var addr string
	if *port == 0 {
		addr = cfg.RUN_SERVER
	} else {
		addr = fmt.Sprintf(":%d", *port)
	}

	log.Printf("[SYSTEM] Proxy Server Up %s", addr)

	server := http.Server{
		Addr:    addr,
		Handler: mux,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Fatal system failure down inside HTTP network router: %v", err)
		}
	}()

	if *origin != "" {
		go func() {
			time.Sleep(time.Millisecond * 100)
			celebUrl := "http://0.0.0.0" + addr + "/proxy?" + "url=" + *origin

			req, err := http.NewRequest("GET", celebUrl, nil)
			if err != nil {
				log.Fatalf("[HTTP] Error triggering warm request %s", err)
				return
			}

			req.Header.Set("X-API", os.Getenv("PROXY_TOKEN"))

			client := &http.Client{}
			res, err := client.Do(req)
			if err != nil {
				log.Fatalf("[HTTP] Error hiting the server %s", err)
				return
			}

			defer res.Body.Close()

			if res.StatusCode == http.StatusOK {
				log.Printf("[HTTP] Cache is fully warmed up and ready")
			} else {
				body, _ := io.ReadAll(res.Body)
				log.Printf("[HTTP] Error unable to cache : status code %d Error Message=%s", res.StatusCode, string(body))
			}
		}()
	}
	<-done

	log.Printf("[SYSTEM] Sutting down the server...")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	_err := server.Shutdown(ctx)

	if _err != nil {
		log.Printf("[SYSTEM] Error Occured while closing the server %s", _err.Error())
	}

	log.Printf("[SYSTEM] Server closed")

}
