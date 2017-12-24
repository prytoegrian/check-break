package check

import (
	"encoding/json"
	"os"
	"strings"
)

type config struct {
	Excluded struct {
		Path string `json:"path"`
	} `json:"excluded"`
}

// loadConfiguration returns a config struct, loaded from parameters
// It doesn't check workingPath validity, as it's already done higher.
func loadConfiguration(workingPath string, configFilename string) *config {
	var conf config
	if !strings.HasSuffix(workingPath, "/") {
		workingPath = workingPath + "/"
	}
	configFilepath := workingPath + configFilename
	configFile, err := os.Open(configFilepath)
	defer configFile.Close()
	if err != nil {
		return nil
	}
	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&conf)
	return &conf
}
