# elastic-apm-daemon (_early alpha_)

_daemon to buffer elastic apm messages and send them as a batch to the apm server._

```
go run daemon.go --url=http://localhost:8000/intake/v2/events
```

options:
```
 --socket=/tmp/.apm.sock
 --url=http://localhost:8000/intake/v2/events
 --send-every=60s
 --buffer=10000
 ```
