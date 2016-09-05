package lib

import (
	"encoding/json"
	"io/ioutil"
)

type conf struct {
	CmdArgs      []string // Command line arguments
	Envs         []string // Environment variables
	Build        string   // Build output
	WatchProcess struct {
		TryRestartOnExit bool // Should we try to rebuild application, if bee catch exit status != 0
	}
	WatchFiles struct {
		Extensions []string // Additional extensions, which bee should watch
		Folders    []string // Additional folders, which bee should watch
	}
	Database struct {
		Driver string
		Conn   string
	}
}
// Config contains all configs, parsed from `bee.json`
var Config conf


// Load config unmarshal bee.json to Config var
func LoadConfig() {
	config, err := ioutil.ReadFile("bee.json")
	if err != nil {
		LogInfo("No bee.json file. Using default configs")
		return
	}
	err = json.Unmarshal(config, &Config)
	LogInfo("Error parsing bee.json file. Using default configs. Error: %v", err)
}
