package bootstrap

import (
	"runtime"

	"github.com/rs/zerolog/log"
)

func Bootstrap() {
	logger()
	config()

	log.Info().
		Str("goarch", runtime.GOARCH).
		Str("goos", runtime.GOOS).
		Str("version", runtime.Version()).
		Msg("Bootstrap successfully")
}
