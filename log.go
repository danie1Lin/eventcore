package eventcore

import (
	"os"

	log "github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

var Logger log.Logger

func init() {
	Logger = log.NewJSONLogger(os.Stdout)
	Logger = log.With(Logger, "ts", log.DefaultTimestampUTC, "caller", log.Caller(5))
}

func DebugMode() {
	Logger = level.NewFilter(Logger, level.AllowDebug())
}
