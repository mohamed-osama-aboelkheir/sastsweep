package logger

import (
	"os"

	"github.com/charmbracelet/log"
)

var logger *log.Logger

func Configure(debug bool) {
	logger = log.New(os.Stderr)
	if debug {
		logger.SetLevel(log.DebugLevel)
	} else {
		logger.SetLevel(log.InfoLevel)
	}
}

func Debug(message string) {
	logger.Debug(message)
}

func Info(message string) {
	logger.Info(message)
}

func Error(message string) {
	logger.Error(message)
}
