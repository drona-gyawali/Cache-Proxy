package http

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/drona-gyawali/cache-proxy/pkg/cache"
	"github.com/drona-gyawali/cache-proxy/pkg/types"
)


type  ProxyServerConfig struct {
	Engine     *cache.LRUCache
	HTTPClient *http.Client
}


func ProxyServerInit (capacity int) *ProxyServerConfig {
	// TODO: make configuration purely dynamic
	customTransport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			KeepAlive: 30 * time.Second,
			Timeout: 5 * time.Second,
		}).DialContext,

		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   10,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   5 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}


	return &ProxyServerConfig{
		Engine: cache.N(capacity),
		HTTPClient: &http.Client{
			Transport: customTransport,
			Timeout: 10 * time.Second,
		},
	}
}


func (P *ProxyServerConfig) ProxyServer (w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	targetUrl := r.URL.Query().Get("url")
	if targetUrl == "" {
		http.Error(w, "The required 'url' paramenter is missing", http.StatusBadRequest)
		return 
	}

	cacheKey := GenerateCacheKey(r, targetUrl)
	if entry, hit := P.Engine.G(cacheKey); hit {
		for headerKey, value := range entry.Headers {
			for _, val := range value {
				w.Header().Add(headerKey, val)
			}
		}

		w.Header().Set("X-Cache", "HIT")
		w.WriteHeader(entry.StatusCode)
		w.Write(entry.Body)

		log.Printf("[CACHE HIT] URL: %s | Duration: %v", targetUrl, time.Since(startTime))
		return
	}


	req , err := http.NewRequestWithContext(r.Context(), http.MethodGet, targetUrl, nil)
	if err != nil {
		http.Error(w, "Request to origin failed", http.StatusBadGateway)
		log.Printf("Origin server request failed via a proxy server, [AFFECTED URL : %s]", targetUrl)
		return 
	}

	copyHeader(r.Header, req.Header)
	req.Header.Del("Connection")
	req.Header.Del("Keep-Alive")

	resp, err := P.HTTPClient.Do(req)
	if err != nil {
		if context.Canceled == err {
			log.Printf("User Disconnect the connection")
			return 
		}

		http.Error(w, "Bad Gateway", http.StatusBadGateway)
		return 
	}
	defer resp.Body.Close()

	bodyBytes, err :=io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to laod bytes into memory")
		return 
	}

	copyHeader(resp.Header, w.Header())
	w.Header().Set("X-Cache", "MISS")

	newEntry := types.CacheEntry{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
		Body:       bodyBytes,
		CachedAt:   time.Now(),
		ExpiresAt:  time.Now().Add(2 * time.Minute),
	}

	P.Engine.S(cacheKey, newEntry)


	w.Header().Del("Cache-Control")
	w.Header().Del("ETag")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	
	w.WriteHeader(resp.StatusCode)
	w.Write(bodyBytes)

	log.Printf("Cached Miss for %s", targetUrl)
}


func copyHeader(src, dst http.Header) {
	for k , vv := range src {
		for _, v := range vv{
			dst.Add(k,v)
		}
	}
}


func GenerateCacheKey(r *http.Request, targetURL string) string {
	authHeader := r.Header.Get("Authorization")
	cookieHeader := r.Header.Get("Cookie")
	customApiKey := r.Header.Get("X-API-Key")

	hasher := sha256.New()
	hasher.Write([]byte(authHeader + "|" + cookieHeader + "|" + customApiKey))
	identityHash := hex.EncodeToString(hasher.Sum(nil))
	log.Printf("[CacheKey] - CrytoGraphic hash been generated")
	return targetURL + ":" + identityHash
}