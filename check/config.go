package check

import (
	"encoding/json"
	"os"
)

type config struct {
	Excluded struct {
		Path string `json:"path"`
	} `json:"excluded"`
}

func loadConfiguration(configPath string) (*config, error) {
	var conf config
	if configPath == "" {
		return &conf, nil
	}
	configFile, err := os.Open(configPath)
	defer configFile.Close()
	if err != nil {
		return nil, err
	}
	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&conf)
	return &conf, nil
}
