package pgrest

import (
	"encoding/json"
	"io"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
)

var config Config
var configFile = getConfigLocation()

type Config struct {
	PGRest      PGRestConfig       `json:"pgrest"`
	Connections []ConnectionConfig `json:"connections"`
}

type PGRestConfig struct {
	Port  int        `json:"port"`
	Debug bool       `json:"debug"`
	CORS  CorsConfig `json:"cors"`
}

type ConnectionConfig struct {
	Name             string       `json:"name"`
	Auth             string       `json:"auth"`
	ConnectionString string       `json:"connectionString"`
	Users            []UserConfig `json:"users"`
}

type UserConfig struct {
	ClientID     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
}

type CorsConfig struct {
	AllowOrigins []string `json:"allowOrigins"`
	AllowHeaders []string `json:"allowHeaders"`
	AllowMethods []string `json:"allowMethods"`
}

// isOriginAllowed checks if the provided origin is allowed based on the CorsConfig settings.
// It returns true if the origin is allowed, otherwise false.
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

func (c CorsConfig) getAllowOriginsString() string {
	return strings.Join(c.AllowOrigins, ", ")
}

func (c CorsConfig) getAllowHeadersString() string {
	return strings.Join(c.AllowHeaders, ", ")
}

func (c CorsConfig) getAllowMethodsString() string {
	return strings.Join(c.AllowMethods, ", ")
}

// getConfigLocation returns the location of the PGREST configuration file.
// If the environment variable PGREST_CONFIG_PATH is set, it returns its value.
// Otherwise, it returns the default location "./config/pgrest.conf".
func getConfigLocation() string {
	location := os.Getenv("PGREST_CONFIG_PATH")
	if location == "" {
		location = "./config/pgrest.conf"
	}
	return location
}

// InitializeConfig loads the configuration and starts watching for changes.
// It returns an error if there was a problem loading the configuration.
func InitializeConfig() error {
	err := loadConfig()
	if err != nil {
		return err
	}

	return nil
}

// loadConfig loads the configuration from a JSON file.
// It reads the JSON file, unmarshals it into the 'config' variable,
// and sets default values if necessary.
// Returns an error if there was a problem reading or unmarshaling the JSON file.
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

	// if debug is not set, default to false
	if !config.PGRest.Debug {
		config.PGRest.Debug = false
	}

	if len(config.PGRest.CORS.AllowOrigins) == 0 {
		config.PGRest.CORS.AllowOrigins = []string{"*"}
	}

	if len(config.PGRest.CORS.AllowHeaders) == 0 {
		config.PGRest.CORS.AllowHeaders = []string{"*"}
	}

	if len(config.PGRest.CORS.AllowMethods) == 0 {
		config.PGRest.CORS.AllowMethods = []string{"POST", "OPTIONS"}
	}

	// iterate over connections and set default values
	for _, conn := range config.Connections {
		if conn.Auth == "" {
			conn.Auth = "private"
		} else if conn.Auth != "public" {
			log.Warnf("Auth for connection '%s' set to public.", conn.Name)
		}
	}

	return nil
}

// GetConfig returns the current configuration.
func GetConfig() Config {
	return config
}
