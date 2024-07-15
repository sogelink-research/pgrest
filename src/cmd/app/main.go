package main

import (
	"os"

	"github.com/sogelink-research/pgrest/server"
	"github.com/sogelink-research/pgrest/settings"

	log "github.com/sirupsen/logrus"
)

func initLogger(config settings.Config) {
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)
	if config.PGRest.Debug {
		log.SetLevel(log.DebugLevel)
	}

	log.SetFormatter(&log.TextFormatter{
		DisableColors: false,
		FullTimestamp: true,
	})
}

func main() {
	err := settings.InitializeConfig()
	if err != nil {
		log.Fatalf("Failed to initialize configuration: %v", err)
	}

	config := settings.GetConfig()
	initLogger(config)
	server.StartServer(config)
}
