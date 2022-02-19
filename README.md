# Requirements

- Golang >= 1.13
- Launched Redis Server >= 5.0

Also, it's required to create **.env** file with the following vars:
- HTTP_PORT - to start app on this port
- REDIS_ADDR - full address to redis server
- REDIS_PASSWORD - password for redis server

For example:
```env
HTTP_PORT=80
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=123ABC
```
