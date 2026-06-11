package observability

import (
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func New(level string) (zerolog.Logger, error) {
	lvl, err := zerolog.ParseLevel(strings.ToLower(level))
	if err != nil {
		lvl = zerolog.InfoLevel
	}
	logger := log.With().Timestamp().Logger().Level(lvl)
	log.Logger = logger
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, NoColor: true})
	return log.Logger.Level(lvl), nil
}
