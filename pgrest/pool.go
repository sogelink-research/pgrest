package pgrest

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	log "github.com/sirupsen/logrus"
)

var (
	dbPoolMap       = make(map[string]*pgxpool.Pool)
	dbPoolMutex     sync.Mutex
	poolLastUsed    = make(map[string]time.Time) // Map to track last usage time of each pool
	cleanupInterval = 1 * time.Minute            // Interval to check for idle pools
)

func init() {
	go periodicCleanup()
}

func periodicCleanup() {
	for {
		time.Sleep(cleanupInterval)

		dbPoolMutex.Lock()
		for name, pool := range dbPoolMap {
			lastUsed, ok := poolLastUsed[name]
			if !ok || time.Since(lastUsed) > cleanupInterval {
				pool.Close()
				delete(dbPoolMap, name)
				delete(poolLastUsed, name)
				log.Debugf("Closed idle pool: %s\n", name)
			}
		}
		dbPoolMutex.Unlock()
	}
}

func CloseDBPools() {
	dbPoolMutex.Lock()
	defer dbPoolMutex.Unlock()
	for _, pool := range dbPoolMap {
		pool.Close()
	}
	dbPoolMap = make(map[string]*pgxpool.Pool)
	poolLastUsed = make(map[string]time.Time)
}

func GetDBPool(name string, connectionString string) (*pgxpool.Pool, error) {
	dbPoolMutex.Lock()
	defer dbPoolMutex.Unlock()

	if pool, ok := dbPoolMap[name]; ok {
		poolLastUsed[name] = time.Now() // Update last used time
		return pool, nil
	}

	pool, err := pgxpool.New(context.Background(), connectionString)
	if err != nil {
		return nil, fmt.Errorf("error connecting to database '%s': %v", name, err)
	}
	dbPoolMap[name] = pool
	poolLastUsed[name] = time.Now() // Update last used time
	return pool, nil
}
