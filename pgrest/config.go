package pgrest

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"strings"
)

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
	Name             string   `json:"name"`
	ConnectionString string   `json:"connectionString"`
	ClientSecrets    []string `json:"clientSecrets"`
}

type CorsConfig struct {
	AllowOrigin  string `json:"allowOrigin"`
	AllowHeaders string `json:"allowHeaders"`
	AllowMethods string `json:"allowMethods"`
}

func (c CorsConfig) getHeaders() []string {
	return strings.Split(c.AllowHeaders, ",")
}

func (c CorsConfig) getMethodsS() []string {
	return strings.Split(c.AllowMethods, ",")
}

func (c CorsConfig) isHeaderAllowed(v string) bool {
	headers := strings.Split(c.AllowHeaders, ",")

	for _, s := range headers {
		if v == s {
			return true
		}
	}
	return false
}

func (c CorsConfig) isMethodAllowed(v string) bool {
	methods := strings.Split(c.AllowMethods, ",")

	for _, s := range methods {
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
