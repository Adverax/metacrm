package bootstrap

import (
	"os"

	"github.com/adverax/metacrm/pkg/enums"
	"github.com/adverax/metacrm/pkg/log"
)

var logLevels = enums.New[log.Level](map[log.Level]string{
	log.TraceLevel: "trace",
	log.DebugLevel: "debug",
	log.InfoLevel:  "info",
	log.WarnLevel:  "warn",
	log.ErrorLevel: "error",
	log.FatalLevel: "fatal",
	log.PanicLevel: "panic",
})

func isDevEnv() bool {
	env := os.Getenv("ENV")
	return env == "dev" || env == "development"
}
