# go-orders-api
I'm currently learning API and Micro-services development, so just made a simple REST api application which implements http Verbs (that is what REST suggests, lol).

In this project, I have used a lightweight and high-performance web framework for go: [chi](https://github.com/go-chi/chi), with [redis](https://github.com/redis/redis) a blazingly fast in-memory data store.

---
For running redis-server I used the following command
> docker run -p 6379:6379 redis:latest

Yeah, docker needs to be installed on your system for this.
