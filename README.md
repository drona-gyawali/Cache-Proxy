> underdevelopment


### Cache-Proxy Server

It acts as an intermediate traffic shield between client browsers and external internet servers. At its core is a custom-engineered LRU (Least Recently Used) cache protected by mutual exclusion locks, ensuring safe data access under heavy, multi-threaded workloads.


### Phase 1: The Arrival & Inspection

1. **Intercept:** A user sends an HTTP request to your proxy (/proxy?url=[https://api.github.com](https://api.github.com)).
2. **Thread Defense:** The proxy instantly calls c.Mu.Lock(). This pauses any other incoming requests trying to modify the cache at the same microsecond.
3. **The Index Lookup:** The proxy checks your CacheMap using the URL string as the search key.

---

### Phase 2: The Decision Split

#### Path A: Cache HIT (Fast Route ~0.5ms)

* **The Expiry Check:** The map finds the item. The proxy instantly checks if time.Now() is past the item's ExpiresAt timestamp.
* **Fresh Data:** If it is still fresh, the engine increments the Hits counter and executes c.EvitList.MoveToFront(node). This snaps the item to the top of your activity timeline.
* **Release & Serve:** The proxy releases the lock (Unlock), injects the X-Cache: HIT header, and dumps the cached body bytes back to the user's browser instantly from RAM.

#### Path B: Cache MISS / STALE (Fetch Route ~200ms)

* **The Clean Up:** If the item wasn't in the map, OR if the timestamp check showed it was expired, it's a miss. If it was expired, the proxy calls c.EvitList.Remove() and delete(c.CacheMap) to clean up memory.
* **Unlock:** The proxy releases the lock early so other threads aren't blocked while we wait for the slow internet.
* **Network Request:** The custom HTTPClient sends an outbound request to the real origin server. It pulls an open, warm connection from your optimized customTransport pool to save time.
* **The Eviction Check:** Once the data returns, the proxy locks again. If your cache length is currently over your max capacity limit, it grabs the oldest node from the absolute back of the list (c.EvitList.Back()) and completely wipes it from memory.
* **Insert & Serve:** The new webpage is packed into a CacheItem, pushed to the front of the timeline list, registered in the map index, and written to the user with X-Cache: MISS.

> <img width="905" height="299" alt="Image" src="https://github.com/user-attachments/assets/2188c69b-7eeb-4dff-8253-aac1becf7fd4" />
