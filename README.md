## Cache-Proxy-Server

A lightweight, high-performance in-memory caching server built from scratch in Go. It acts as an intermediate accelerator specifically designed for **REST API requests**, slashing backend response latencies down to microseconds while shielding origin servers from heavy traffic.

> <img width="905" height="299" alt="Image" src="https://github.com/user-attachments/assets/2188c69b-7eeb-4dff-8253-aac1becf7fd4" />


### Key Features

* **Built for APIs Only:** Optimizes JSON/REST API responses, reducing subsequent request latencies from >50ms to under 1ms.
  
* **Thread-Safe LRU Cache:** Custom-engineered Least Recently Used (LRU) eviction policy protected by mutual exclusion (`sync.Mutex`) locks for safe concurrent operations.
  
* **Domain Whitelisting:** Enhances security by only forwarding requests to external domains pre-registered in `config/local.yaml`.
  
* **Token Authentication:** Secure proxy access via required `X-API` headers.

---

### How to Implement It (Quick Start)

> https://github.com/user-attachments/assets/e568a4a8-5f24-4394-a473-8b2fc1335b80


1. **Clone the repository:**
```bash
git clone https://github.com/drona-gyawali/Cache-Proxy.git
cd Cache-Proxy

```


2. **Configure your environment:**
Copy `.env.example` to `.env` and set your secret token. Register your allowed external API domains in `config/local.yaml`.
3. **Boot up the server:**
```bash
go run cmd/cache_proxy/main.go

```
4. **Warm the Cache**
```bash
go run cmd/cache_proxy/main.go --port 8080 --origin https://dummyjson.com/comments

```

4. **Test the API Cache:**
```bash
curl -i \
  -H "X-API: YOUR_TOKEN_HERE" \
  "http://localhost:8080/proxy?url=https://dummyjson.com/comments"

```


*(Note: The first request will fetch from the origin server, while the second will load instantly from the memory cache.)*
