package settings

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
)

var config Config
var configFile = getConfigLocation()

type Config struct {
	PGRest      PGRestConfig       `json:"pgrest"`
	Connections []ConnectionConfig `json:"connections"`
	Users       []UserConfig       `json:"users"`
	UsersLookup map[string]UserConfig
}

// getConnectionConfig retrieves the connection configuration for the given name.
// It searches through the list of connections in the Config object and returns
// the ConnectionConfig if found, otherwise it returns an error.
func (c Config) GetConnectionConfig(name string) (*ConnectionConfig, error) {
	for _, conn := range c.Connections {
		if conn.Name == name {
			return &conn, nil
		}
	}

	return nil, fmt.Errorf("connection %s not found", name)
}

type PGRestConfig struct {
	Port                  int        `json:"port"`
	Debug                 bool       `json:"debug"`
	CORS                  CorsConfig `json:"cors"`
	MaxConcurrentRequests int        `json:"maxConcurrentRequests"`
	Timeout               int        `json:"timeout"`
}

type ConnectionConfig struct {
	Name             string `json:"name"`
	Auth             string `json:"auth"`
	ConnectionString string `json:"connectionString"`
}

type UserConfig struct {
	ClientID     string   `json:"clientId"`
	ClientSecret string   `json:"clientSecret"`
	Connections  []string `json:"connections"`
}

type CorsConfig struct {
	AllowOrigins []string `json:"allowOrigins"`
	AllowHeaders []string `json:"allowHeaders"`
	AllowMethods []string `json:"allowMethods"`
}

// isOriginAllowed checks if the provided origin is allowed based on the CorsConfig settings.
// It returns true if the origin is allowed, otherwise false.
func (c CorsConfig) IsOriginAllowed(v string) bool {
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

func (c CorsConfig) GetAllowOriginsString() string {
	return strings.Join(c.AllowOrigins, ", ")
}

func (c CorsConfig) GetAllowHeadersString() string {
	return strings.Join(c.AllowHeaders, ", ")
}

func (c CorsConfig) GetAllowMethodsString() string {
	return strings.Join(c.AllowMethods, ", ")
}

// getConfigLocation returns the location of the PGREST configuration file.
// If the environment variable PGREST_CONFIG_PATH is set, it returns its value.
// Otherwise, it returns the default location "./config/pgrest.conf".
func getConfigLocation() string {
	location := os.Getenv("PGREST_CONFIG_PATH")
	if location == "" {
		location = "../config/pgrest.conf"
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

	// Preprocess the JSON to remove excessive commas
	cleanedJSON := cleanJSON(string(byteValue))

	err = json.Unmarshal([]byte(cleanedJSON), &config)
	if err != nil {
		return err
	}

	if config.PGRest.Port == 0 {
		config.PGRest.Port = 8080
	}

	if config.PGRest.MaxConcurrentRequests == 0 {
		config.PGRest.MaxConcurrentRequests = 15
	}

	if config.PGRest.Timeout == 0 {
		config.PGRest.Timeout = 30
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

	config.UsersLookup = make(map[string]UserConfig)
	for _, user := range config.Users {
		config.UsersLookup[user.ClientID] = user
	}

	return nil
}

func cleanJSON(input string) string {
	// Remove trailing commas before closing braces and brackets
	re := regexp.MustCompile(`,\s*([\]}])`)
	cleaned := re.ReplaceAllString(input, "$1")
	// Ensure that there are no consecutive commas
	cleaned = strings.ReplaceAll(cleaned, ",,", ",")
	return cleaned
}

// GetConfig returns the current configuration.
func GetConfig() Config {
	return config
}
