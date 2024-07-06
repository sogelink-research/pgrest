package pgrest

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
)

var config Config
var configFile = "./config/pgrest.conf"

type Config struct {
	PGRest      PGRestConfig       `json:"pgrest"`
	Connections []ConnectionConfig `json:"connections"`
}

type PGRestConfig struct {
	Port  int  `json:"port"`
	Debug bool `json:"debug"`
}

type ConnectionConfig struct {
	Name             string       `json:"name"`
	ConnectionString string       `json:"connectionString"`
	Users            []UserConfig `json:"users"`
}

type UserConfig struct {
	ClientID     string     `json:"clientId"`
	APIKey       string     `json:"apiKey"`
	ClientSecret string     `json:"clientSecret"`
	CORS         CorsConfig `json:"cors"`
}

type CorsConfig struct {
	AllowOrigins []string `json:"allowOrigins"`
}

func (c CorsConfig) isOriginAllowed(v string) bool {
	if c.AllowOrigins[0] == "*" {
		return true
	}

	for _, s := range c.AllowOrigins {
		if v == s {
			return true
		}
	}
	return false
}

func InitializeConfig() error {
	err := loadConfig()
	if err != nil {
		return err
	}

	go watchConfigChanges()

	return nil
}

// loadConfig reads and parses the config file.
func loadConfig() error {
	jsonFile, err := os.Open(configFile)
	if err != nil {
		return err
	}
	defer jsonFile.Close()

	byteValue, err := io.ReadAll(jsonFile)
	if err != nil {
		return err
	}

	err = json.Unmarshal(byteValue, &config)
	if err != nil {
		return err
	}

	// Set default values if necessary
	if config.PGRest.Port == 0 {
		config.PGRest.Port = 8080
	}

	return nil
}

func GetConfig() Config {
	return config
}

func watchConfigChanges() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatalf("Failed to create file watcher: %v", err)
	}
	defer watcher.Close()

	configDir := filepath.Dir(configFile)

	err = watcher.Add(configDir)
	if err != nil {
		log.Fatalf("Failed to watch config file changes: %v", err)
	}

	log.Debugf("Watching for changes in %s", configDir)

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				log.Debugf("Config file changed. Reloading...")
				time.Sleep(100 * time.Millisecond) // Add a small delay to handle rapid changes
				err := loadConfig()
				if err != nil {
					log.Errorf("Error reloading config: %v", err)
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Errorf("Error watching config file: %v", err)
		}
	}
}
