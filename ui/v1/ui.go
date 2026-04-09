package v1

import (
	"github.com/dubeyKartikay/lazyspotify/core/logger"
	"github.com/dubeyKartikay/lazyspotify/ui/v1/app"
)

func RunTui() {
	if err := app.Run(); err != nil {
		logger.Log.Error().Err(err).Msg("failed to run program")
	}
}
