// File: setup/setupLogger.go
package setup

import (
	"dingus/pkg/config"
	"os"

	loki "github.com/saromanov/logrus-loki-hook"
	"github.com/sirupsen/logrus"
)

// TODO: If we load configuration in main.go before the logger we can add Loki hook URL to the configuration
func SetupLogger(cfg *config.Config) *logrus.Logger {
	log := logrus.New()

	setLokiHook(log, cfg)
	log.SetFormatter(&logrus.JSONFormatter{})

	logLevel, err := logrus.ParseLevel(os.Getenv("LOG_LEVEL"))
	if err != nil {
		logLevel = logrus.InfoLevel
	}
	log.SetLevel(logLevel)

	return log
}

func setLokiHook(log *logrus.Logger, cfg *config.Config) {
	hook, err := loki.NewHook(&loki.Config{
		URL: cfg.LokiHookURL,
	})
	if err != nil {
		log.Error("Loki hook initialization failed")
	} else {
		log.AddHook(hook)
	}
}
