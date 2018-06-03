// Handles all the configuration for Xerophi Admin
package main

import (
	"encoding/json"
	"os"

	"github.com/CactusDev/Xerophi/redis"
	"github.com/CactusDev/Xerophi/rethink"
)

// Config keeps track of the config set in config.json
type Config struct {
	Rethink rethinkCfg `json:"rethink"`
	Sentry  sentryCfg  `json:"sentry"`
	Server  serverCfg  `json:"server"`
	Redis   redisCfg   `json:"redis"`
}

type rethinkCfg struct {
	Connection rethink.ConnectionOpts `json:"connection"`
	DB         string                 `json:"db"`
}

type sentryCfg struct {
	DSN     string `json:"dsn"`
	Enabled bool   `json:"enabled"`
}

type redisCfg struct {
	Connection redis.ConnectionOpts `json:"connection"`
	DB         int                  `json:"db"`
}

type serverCfg struct {
	Port int `json:"port"`
}

// LoadConfigFromPath loads config from a specific path
func LoadConfigFromPath(path string) Config {
	config := Config{}
	configFile, err := os.Open(path)
	defer configFile.Close()
	if err != nil {
		panic(err)
	}

	jsonParser := json.NewDecoder(configFile)
	if err = jsonParser.Decode(&config); err != nil {
		panic(err)
	}

	return config
}
