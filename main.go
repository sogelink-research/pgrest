package main

import (
	"fmt"
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
	config := pgrest.GetConfig()
	initLogger(config)
	log.Info(fmt.Sprintf("PGRest started, running on port %v", config.PGRest.Port))
	pgrest.StartServer(config)
}
