## go-redis

Implementation of simple redis like KV store in Golang. Inspired from [Build your Own Redis](https://build-your-own.org/redis/). This only supports BSD based distros. Support for UNIX is yet to be added.

Note: This is just a small side project that I created to learn more about databases. This is no way close to solving production usecases.

## Setup

### Server

- Install Go 1.20
- Clone this repo
- Run `make dev` to build the code and generate the binary.
- The binary will be placed in the `dist/` folder in the root directory. You can either add the binary to `$PATH` or invoke it directly by `./dist/darwin/arm64/go-redis`.
- This should startup the server.

### Client

- The server uses the [RESP protocol](https://redis.io/docs/reference/protocol-spec/) to listen/respond to requests. Any Redis client can be used to talk to it. Following shows an example with `redis-cli`

```shell
redis_go % redis-cli -p 1234
127.0.0.1:1234> GET a
(error) key not found
127.0.0.1:1234> SET a 1234
1) OK
127.0.0.1:1234> GET a
1) "1234"
127.0.0.1:1234> DEL a
1) OK
127.0.0.1:1234> GET a
(error) key not found
```

## Features

- Supports handling concurrent client connections with event loops and non blocking I/O.
- Supports pipelined client requests.
- Supports parsing requests with RESP protocol.
- Supports only simple commands like GET/SET/DEL for now.
- The backend store is a simple Golang map. This will be replaced with hashtables in the upcoming iterations.

## Gaps

- Collision handling
- Transactions
- Key expiry and eviction
- Efficient storage and retrieval of keys
