package http

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/drona-gyawali/cache-proxy/pkg/cache"
	"github.com/drona-gyawali/cache-proxy/pkg/types"
)

type ProxyServerConfig struct {
	Engine         *cache.LRUCache
	HTTPClient     *http.Client
	ProxyToken     string
	AllowedOrigins map[string]string
	CacheQueue     chan types.CacheItem
}

func ProxyServerInit(capacity int, proxyToken string, allowedOrigins map[string]string) *ProxyServerConfig {
	// TODO: make configuration purely dynamic
	customTransport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			KeepAlive: 30 * time.Second,
			Timeout:   5 * time.Second,
		}).DialContext,

		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   10,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   5 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	config := &ProxyServerConfig{
		Engine: cache.N(capacity),
		HTTPClient: &http.Client{
			Transport: customTransport,
			Timeout:   10 * time.Second,
		},
		ProxyToken:     proxyToken,
		AllowedOrigins: allowedOrigins,
		CacheQueue:     make(chan types.CacheItem, 1000),
	}

	// this kind of async queue processing is fancy but i am afraid that this might create a problem.
	// user 1 and user 2 wants to cache a same url because we have async which has its own latencies
	// might choke. I am dubious this is not scalable in prod.. however keeping it as it is.
	go func() {
		log.Printf("[SYSTEM] Async Cache Processing Worker Started.")
		for item := range config.CacheQueue {
			config.Engine.S(item.Key, item.Value)
		}
	}()

	return config
}

func (P *ProxyServerConfig) ProxyServer(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	remoteAddr := r.Host

	isVerified := P.VerifiedServerOrigins(w, remoteAddr)
	if !isVerified {
		return
	}
	clientToken := r.Header.Get("X-API")
	if clientToken == "" || subtle.ConstantTimeCompare([]byte(clientToken), []byte(P.ProxyToken)) != 1 {
		log.Printf("[SECURITY] Invalid token used by server IP")
		http.Error(w, "Unauthorized: Invalid Proxy Cluster Token", http.StatusUnauthorized)
		return
	}

	targetUrl := r.URL.Query().Get("url")
	if targetUrl == "" {
		log.Fatalf("The Required %s 'url' parameter is missing", targetUrl)
		http.Error(w, "The required 'url' parameter is missing", http.StatusBadRequest)
		return
	}
	parsedUrl, err := url.Parse(targetUrl)
	if err != nil || parsedUrl.Scheme == "" || parsedUrl.Host == "" {
		http.Error(w, "Invalid target URL format", http.StatusBadRequest)
		return
	}
	originServer := parsedUrl.Scheme + "://" + parsedUrl.Host
	originServer = strings.TrimSuffix(originServer, "/")
	if !P.VerifiedServerOrigins(w, originServer) {
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

	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, targetUrl, nil)
	if err != nil {
		http.Error(w, "Request to origin failed", http.StatusBadGateway)
		log.Printf("Origin server request failed via a proxy server, [AFFECTED URL : %s]", targetUrl)
		return
	}

	copyHeader(r.Header, req.Header)
	req.Header.Del("Connection")
	req.Header.Del("Keep-Alive")

	req.Header.Del("X-API")
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

	bodyBytes, err := io.ReadAll(resp.Body)
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
		ExpiresAt:  time.Now().Add(5 * time.Minute),
	}

	cacheJob := types.CacheItem{
		Key:   cacheKey,
		Value: newEntry,
	}

	select {
	case P.CacheQueue <- cacheJob:
		log.Printf("[SYSTEM] Cache Write Queued")
	default:
		log.Printf("[SYSTEM] Cache Queue Failed")
	}

	w.Header().Del("Cache-Control")
	w.Header().Del("ETag")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")

	w.WriteHeader(resp.StatusCode)
	w.Write(bodyBytes)

	log.Printf("Cached Miss for %s", targetUrl)
}

func copyHeader(src, dst http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

// TODO: this is a heavy cpu intensive task we have to make algo very light to increase latency :)
func GenerateCacheKey(r *http.Request, targetURL string) string {
	authHeader := r.Header.Get("Authorization")
	cookieHeader := r.Header.Get("Cookie")
	customApiKey := r.Header.Get("X-API-Key")

	// if it is global then the cache should be also global :)
	if authHeader == "" && cookieHeader == "" && customApiKey == "" {
		return targetURL
	}

	hasher := sha256.New()
	hasher.Write([]byte(authHeader + "|" + cookieHeader + "|" + customApiKey))
	identityHash := hex.EncodeToString(hasher.Sum(nil))
	log.Printf("[CacheKey] - CrytoGraphic hash been generated")
	return targetURL + ":" + identityHash
}

func (P *ProxyServerConfig) VerifiedServerOrigins(w http.ResponseWriter, remoteAddr string) bool {
	allowed := false
	for _, v := range P.AllowedOrigins {
		if v == remoteAddr {
			allowed = true
			break
		}
	}

	if !allowed {
		error := "Access Blocked: Destination Server is not registered: " + remoteAddr
		log.Printf(error)
		http.Error(w, error, http.StatusUnauthorized)

		return false
	}

	return true
}
