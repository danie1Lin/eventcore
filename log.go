package eventcore

import (
	"os"

	log "github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

var Logger log.Logger
var Debug bool

func init() {
	Debug = true
	Logger = log.NewJSONLogger(os.Stdout)
	Logger = log.With(Logger, "ts", log.DefaultTimestampUTC, "caller", log.Caller(5))
	if Debug {
		Logger = level.NewFilter(Logger, level.AllowDebug())
	} else {
		Logger = level.NewFilter(Logger, level.AllowInfo())
	}
}
