> underdevelopment


### Cache-Proxy Server
THe idea is to built single web server pointed to a cache proxy server to measure latency savings, cache hit ratios, and evaluate cache eviction policies.

### Request Lifecycle

1. User Request Path: The client issues an `HTTP GET /content` request directed at the Cache Proxy Server (`192.168.1.10`).

2. Hit/Miss Evaluation: 
   - Cache Hit: The requested resource exists in **Local Cache Storage**. The proxy immediately serves the asset via the **Cache Hit Path**, bypassing the origin.

   - Cache Miss: The proxy intercepts the request, logs a fetch miss, and forwards the path to the **Origin Webserver** (`192.168.1.20`) to retrieve and store the fresh asset.

   - Eviction Management: When local storage limits are reached, the system triggers the active **Eviction Policy** (e.g., LRU) to clear stale data.


> <img width="905" height="299" alt="Image" src="https://github.com/user-attachments/assets/2188c69b-7eeb-4dff-8253-aac1becf7fd4" />