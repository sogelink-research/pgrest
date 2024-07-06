package main

import (
	"os"

	"github.com/sogelink-research/pgrest/pgrest"

	log "github.com/sirupsen/logrus"
)

func initLogger(config pgrest.Config) {
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
	pgrest.InitializeConfig()
	config := pgrest.GetConfig()
	initLogger(config)
	pgrest.StartServer(config)
}
