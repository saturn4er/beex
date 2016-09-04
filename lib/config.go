package lib

import (
	"encoding/json"
	"io/ioutil"
)

type conf struct {
	WatchFiles struct {
		Extensions []string
		Folders    []string
	}
	WatchProcess struct {
		TryRestartOnExit bool
	}
	CmdArgs  []string
	Envs     []string
	Database struct {
		Driver string
		Conn   string
	}
}

var Config conf

func LoadConfig() {
	config, err := ioutil.ReadFile("bee.json")
	if err != nil {
		LogInfo("No bee.json file. Using default configs")
		return
	}
	err = json.Unmarshal(config, &Config)
	LogInfo("Error parsing bee.json file. Using default configs. Error: %v", err)
}
