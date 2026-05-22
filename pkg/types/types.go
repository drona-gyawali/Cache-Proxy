package types


import (
	"time"
)

type CacheEntry struct {
	StatusCode 		int
    Headers    		map[string][]string
    Body       		[]byte
    CachedAt   		time.Time
    ExpiresAt  		time.Time
    Hits       		int
    ETag           	string
    LastModified   	string

    ContentEncoding string
}

type  CacheItem struct {
	Key string
	Value CacheEntry
}


