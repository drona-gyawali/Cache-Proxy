package cache

import (
	"container/list"
	"log"
	"sync"
	"time"

	"github.com/drona-gyawali/cache-proxy/pkg/types"
)


const (
	DEFAULT_CAP = 100
)


type LRUCache struct {
	Mu sync.Mutex
	Capacity int
	EvitList *list.List
	CacheMap  map[string]*list.Element
}



func N (capacity int) *LRUCache {
	if capacity < 5 {capacity = DEFAULT_CAP}
	return &LRUCache{
		Capacity : capacity,
		EvitList : list.New(),
		CacheMap : make(map[string]*list.Element),
	}
}


func (C *LRUCache) G  (key string) (types.CacheEntry, bool) {
	C.Mu.Lock()
	defer C.Mu.Unlock()

	node, exist := C.CacheMap[key]
	if !exist {
		log.Printf("[GET]: Targeted Address MISS %s", key)
		return types.CacheEntry{}, false
	}

	items := node.Value.(*types.CacheItem)
	entry := items.Value

	if time.Now().After(entry.ExpiresAt) {
		delete(C.CacheMap, key)
		C.EvitList.Remove(node)
		log.Printf("[GET]: Targeted Address Expired %s", key)
		return types.CacheEntry{}, false
	}

	entry.Hits++
	items.Value = entry
	C.EvitList.MoveToFront(node)
	log.Printf("[GET] Target Address Hit %s", key)
	return entry, true
}


func (C *LRUCache) S (key string, entry types.CacheEntry) () {
	C.Mu.Lock()
	defer C.Mu.Unlock()

	if node , exists := C.CacheMap[key];exists {
		entryItems := node.Value.(*types.CacheItem)
		entryItems.Value =  entry
		C.EvitList.MoveToFront(node)
		log.Printf("[SET] Targeted Address Updated %s", key)
		return
	}

	if C.EvitList.Len() >= C.Capacity {
		unusedItems := C.EvitList.Back()
		unusedItemsVal := unusedItems.Value.(*types.CacheItem)
		if unusedItems != nil {
			delete(C.CacheMap, unusedItemsVal.Key)
			C.EvitList.Remove(unusedItems)
			log.Printf("[SET] Targeted Address Capacity Reached %s", unusedItemsVal.Key)
		}
	}


	newItem := &types.CacheItem{
		Key:key,
		Value: entry,
	}

	newNode := C.EvitList.PushFront(newItem)
	C.CacheMap[key] = newNode
	log.Printf("[SET] Targeted Address Created %s", key)
}