# patroniglue
> Handle and cache basic Patroni API checks

[Patroni](https://github.com/zalando/patroni) uses the built-in Python HTTP server to expose database states. It's perfect to be used by a load balancer like HAProxy to achieve high-availability. But, sometimes, this interface freezes. There's an [open issue](https://github.com/zalando/patroni/issues/857) we are trying to close actively. As production doesn't wait, `patroniglue` was created to offload those checks and release pressure by adding a little response cache.

## Usage
Start process using a configuration file:
```
patroniglue -config config.yml
```
Add more logging output:
```
patroniglue -config config.yml -verbose
```
Print usage:
```
patroniglue -help
```

## Configuration

Configuration file format is YAML.

* `frontend`: settings to handle incoming requests
  * `host`: address to handle requests (localhost by default)
  * `port`: port to handle requests (80 by default)
  * `certfile`: path to SSL certificate file (will use HTTP by default if not provided)
  * `keyfile`: path to SSL private key file (will use HTTP by default if not provided)
  * `tls-min-version`: minimum TLS version for HTTPS service (could be `SSLv3.0`, `TLSv1.0`,`TLSv1.1`, `TLSv1.2`, `TLSv1.3`)
  * `tls-ciphers`: list of supported ciphers (see [full list](https://golang.org/pkg/crypto/tls/#pkg-constants))
* `backend`: settings for sending requests to a backend
  * `host`: patroni REST API `listen` address
  * `port`: patroni REST API `listen` port
  * `scheme`: patroni REST API scheme (either `http` or `https`)
  * `insecure`: disable certificate checks on HTTPS requests
* `cache`: settings for the caching system
  * `ttl`: time in second before response will be evinced
  * `interval`: time in second used by the internal cache loop to check for keys to remove

See [config.yml.example](config.yml.example) file for an example.

## Internals

* Frontend handles HTTP or HTTPS requests on "/primary", "/master", "/replica", "/read-write" and "/read-only" routes available on Patroni API
* Backend requests Patroni API using HTTP or HTTPS protocol and exposes state to frontend
* Cache implements an in-memory key-value store to cache backend responses for some time

## Build

Run `./build.sh` script and enjoy!
