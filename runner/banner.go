package runner

import (
	"fmt"
	"sastsweep/common/logger"
)

const banner = `  ___   _   ___ _____                       
 / __| /_\ / __|_   _____ __ _____ ___ _ __ 
 \__ \/ _ \\__ \ | |(_-\ V  V / -_/ -_| '_ \
 |___/_/ \_|___/ |_|/__/\_/\_/\___\___| .__/
                                      |_|`

const version = "v0.0.1"

func ShowBanner(noEmoji bool) {
	fmt.Println(banner)

	logger.Info("Current SASTsweep version: " + version)
	if noEmoji {
		logger.Info("Made by @_chebuya with <3")
	} else {
		logger.Info("Made by @_chebuya with 🩷")
	}
}
