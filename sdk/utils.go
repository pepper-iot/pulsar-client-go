package sdk

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.elastic.co/ecszerolog"
)

func init() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	logger := ecszerolog.New(os.Stdout)
	log.Logger = logger

	if _, ok := os.LookupEnv("LOCAL_LOGGER"); ok {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout}) // pretty print to stderr
		log.Logger = log.With().Caller().Logger()                      // shows filename and line number
	}

	if level, ok := os.LookupEnv("LOG_LEVEL"); ok {
		switch level {
		case "debug":
			zerolog.SetGlobalLevel(zerolog.DebugLevel)
		case "info":
			zerolog.SetGlobalLevel(zerolog.InfoLevel)
		case "warn":
			zerolog.SetGlobalLevel(zerolog.WarnLevel)
		case "error":
			zerolog.SetGlobalLevel(zerolog.ErrorLevel)
		case "panic":
			zerolog.SetGlobalLevel(zerolog.PanicLevel)
		default:
			logger.Error().Msgf("unknown value for LOG_LEVEL: %s, using debug", level)
			zerolog.SetGlobalLevel(zerolog.DebugLevel)
		}
	} else {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}
}
