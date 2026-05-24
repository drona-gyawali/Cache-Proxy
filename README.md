### Cache-Proxy Server

> <img width="905" height="299" alt="Image" src="https://github.com/user-attachments/assets/2188c69b-7eeb-4dff-8253-aac1becf7fd4" />

It acts as an intermediate traffic shield between client browsers and external internet servers. At its core is a custom-engineered LRU (Least Recently Used) cache protected by mutual exclusion locks, ensuring safe data access under heavy, multi-threaded workloads.

> https://github.com/user-attachments/assets/e568a4a8-5f24-4394-a473-8b2fc1335b80

### How it works
```
git clone https://github.com/drona-gyawali/Cache-Proxy.git

// this will boot up the server
go run cmd/cache_proxy/main.go


```
> **Note: Proxy server will required to have a secret token attach in every request as shown below, for more info please refer to .env.example**

> **Open console and hit below requests intially it takes around > 50ms**

> **Next time hit the same request it will load with in a less than  797.308µs**

```
curl -i \
  -H "X-API: YOUR_TOKEN_HERE" \
  "http://localhost:8080/proxy?url=https://dummyjson.com/comments"

```

