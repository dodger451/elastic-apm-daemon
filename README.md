# elastic-apm-daemon (_early alpha_)

_daemon to buffer elastic apm messages and send them as a batch to the apm server._

## Usage


### use with apm v1  api

```
 go run daemon.go --url=http://0.0.0.0:8200/v1/transactions  --send-every=5s

 go run daemon.go --url=http://0.0.0.0:8200/v1/errors  --send-every=5s --socket=/tmp/.apm_error.sock

```


### use with apm v2 intake api

```
go run daemon.go --url=http://localhost:8000/intake/v2/events --content-type-header=application/x-ndjson
```

### options
```
 --socket=/tmp/.apm.sock
 --url=http://localhost:8000/intake/v2/events
 --send-every=60s
 --buffer=10000
 --content-type-header=application/json
 ```
