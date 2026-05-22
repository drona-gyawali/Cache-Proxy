package http

import (
	"net"
	"net/http"
	"time"

	"github.com/drona-gyawali/cache-proxy/pkg/cache"
)

// first layer : Network and proxy controller
// second layer: Optimized Client Transport


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


// TODO: 
func ProxyServer () {
	
}