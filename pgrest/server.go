package pgrest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"time"

	log "github.com/sirupsen/logrus"
)

func StartServer(config Config) {
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		var m runtime.MemStats
		for range ticker.C {
			runtime.ReadMemStats(&m)
			fmt.Printf("Used Memory = %v MiB\n", m.Alloc/1024/1024)
		}
	}()

	http.Handle("/", middleware(QueryHandler, config.Connections))
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", config.PGRest.Port), nil))
}

func middleware(handler func(http.ResponseWriter, *http.Request, ConnectionConfig, RequestBody), connections []ConnectionConfig) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var requestBody RequestBody
		if err := parseRequestBody(r, &requestBody); err != nil {
			http.Error(w, "Failed to parse JSON", http.StatusBadRequest)
			return
		}

		// in connections find the one with the same name as the database in the request
		var connection ConnectionConfig
		for _, c := range connections {
			if c.Name == requestBody.Database {
				connection = c
				break
			}
		}

		// if no connection is found, return an error
		if connection.Name == "" {
			http.Error(w, "Database not found", http.StatusNotFound)
			return
		}

		handler(w, r, connection, requestBody)
	})
}

func parseRequestBody(req *http.Request, body *RequestBody) error {
	return json.NewDecoder(req.Body).Decode(body)
}
