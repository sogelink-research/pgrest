package pgrest

import (
	"encoding/json"
	"io"
	"log"
	"os"
)

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

func GetConfig() Config {
	jsonFile, err := os.Open("./config/pgrest.conf")
	if err != nil {
		log.Fatalf("Unable to open config: %v", err.Error())
	}

	defer jsonFile.Close()
	byteValue, _ := io.ReadAll(jsonFile)
	var config Config
	err = json.Unmarshal(byteValue, &config)

	if err != nil {
		log.Fatalf("Unable to parse config: %v", err.Error())
	}

	if config.PGRest.Port == 0 {
		config.PGRest.Port = 8080
	}

	return config
}
