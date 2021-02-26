# Theseus
![img](https://img.shields.io/docker/cloud/build/decentr/theseus.svg) ![img](https://img.shields.io/github/go-mod/go-version/Decentr-net/theseus) ![img](https://img.shields.io/github/v/tag/Decentr-net/theseus?label=version)

Theseus provides Decentr community off-chain functionality.

## Run
### Import genesis to database
```
go run scripts/genesis2db/main.go --genesis.json /path/to/genesis.json --postgres "host=localhost port=5432 user=postgres password=root sslmode=disable" --postgres.migrations "scripts/migrations/postgres"
```

### Docker
#### Local image
```
make image
docker run -it --rm -e "HTTP_HOST=0.0.0.0" -e "HTTP_PORT=7070" -e "LOG_LEVEL=debug" -p "7080:7070" theseus-local
```
### From source
```
go run cmd/theseus/main.go \
    --http.host=0.0.0.0 \
    --http.port=8080 \
    --log.level=debug \
    --postgres="host=localhost port=5432 user=postgres password=root sslmode=disable" \
    --postgres.migrations="scripts/migrations/postgres" \
    --blockchain.node="zeus.testnet.decentr.xyz:26656" \
```

## Parameters

| CLI param         | Environment var          | Default | Required | Description
|---------------|------------------|---------------|-------|---------------------------------
| http.host         | HTTP_HOST         | 0.0.0.0  | true | host to bind server
| http.port    | HTTP_PORT    | 8080  | true | port to listen
| postgres    | POSTGRES    | host=localhost port=5432 user=postgres password=root sslmode=disable  | true | postgres dsn
| postgres.max_open_connections    | POSTGRES_MAX_OPEN_CONNECTIONS    | 0 | true | postgres maximal open connections count, 0 means unlimited
| postgres.max_idle_connections    | POSTGRES_MAX_IDLE_CONNECTIONS    | 5 | true | postgres maximal idle connections count
| postgres.migrations    | POSTGRES_MIGRATIONS    | /migrations/postgres | true | postgres migrations directory
| blockchain.node   | BLOCKCHAIN_NODE    | http://zeus.testnet.decentr.xyz:26657 | true | decentr node address
| blockchain.timeout   | BLOCKCHAIN_TIMEOUT    | 5s| true | timeout for requests to blockchain node
| blockchain.retry_interval   | BLOCKCHAIN_RETRY_INTERVAL    | 2s | true | interval to be waited on error before retry
| blockchain.last_block_retry_interval   | BLOCKCHAIN_LAST_BLOCK_RETRY_INTERVAL    | 1s | true | duration to be waited when new block isn't produced before retry
| log.level   | LOG_LEVEL   | info | false | level of logger (debug,info,warn,error)
| sentry.dsn    | SENTRY_DSN    |  | false | sentry dsn

## Development
### Makefile
#### Update vendors
Use `make vendor`
#### Install required for development tools
You can check all tools existence with `make check-all` or force installing them with `make install-all` 
##### golangci-lint 1.29.0
Use `make install-linter`
##### swagger v0.25.0
Use `make install-swagger`
##### gomock v1.4.3
Use `make install-mockgen`
#### Build docker image
Use `make image` to build local docker image named `theseus-local`
#### Build binary
Use `make build` to build for your OS or use `make linux` to build for linux(used in `make image`) 
#### Run tests
Use `make test` to run tests. Also you can run tests with `integration` tag with `make fulltest`
