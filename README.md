[![GoDoc](https://godoc.org/github.com/bakins/twirpzap?status.svg)](https://godoc.org/github.com/bakins/twirpzap)

# twirp zap logger

Logger for [twirp](https://twitchtv.github.io/twirp/docs/intro.html) servers using [zap](https://github.com/uber-go/zap)

## Usage

Install locally: `go get -u github.com/bakins/twirpzap`

Create server hooks:

```go
import (
	"github.com/twitchtv/twirp/example"
    "go.uber.org/zap"
    "github.com/bakins/twirpzap"
)

func main() {
    logger, _ := zap.NewProduction()
    defer logger.Sync() 

    server := example.NewHaberdasherServer(&testHaberdasher{}, twirpzap.ServerHooks(logger))	
}
```

Log lines will look like:

```
{"level":"info","ts":1557966347.879602,"caller":"twirp-zap-logger/logger.go:62","msg":"response sent","twirp.package":"twitch.twirp.example","twirp.service":"Haberdasher","twirp.method":"MakeHat","twirp.status":"200","duration":0.000169998}
```

See also [./example/server.go](./example/server.go)

## LICENSE

See [LICENSE](./LICENSE)


