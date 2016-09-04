package lib

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"
)

// GetGOPATHs returns all paths in GOPATH variable.
func GetGOPATHs() []string {
	gopath := os.Getenv("GOPATH")
	var paths []string
	if runtime.GOOS == "windows" {
		gopath = strings.Replace(gopath, "\\", "/", -1)
		paths = strings.Split(gopath, ";")
	} else {
		paths = strings.Split(gopath, ":")
	}
	return paths
}

func getSubDirectories(dirPath string) ([]string, error) {
	fileInfos, err := ioutil.ReadDir(dirPath)
	result := []string{}
	if err != nil {
		return result, err
	}
	useDirectory := false
	for _, fileInfo := range fileInfos {
		if strings.HasSuffix(fileInfo.Name(), "docs") {
			continue
		}

		if fileInfo.IsDir() == true && fileInfo.Name()[0] != '.' {
			add, err := getSubDirectories(dirPath + "/" + fileInfo.Name())
			if err != nil {
				return result, err
			}
			result = append(result, add...)
			continue
		}

		if useDirectory == true {
			continue
		}

		if path.Ext(fileInfo.Name()) == ".go" {
			result = append(result, dirPath)
			useDirectory = true
		}
	}
	return result, nil
}

func GetPathDeps(path string) ([]string, error) {
	icmd := exec.Command("go", "list", "-f", "{{.Deps}}", path+"/...")
	buf := &bytes.Buffer{}
	errorB := &bytes.Buffer{}
	icmd.Stdout = buf
	icmd.Stderr = errorB
	icmd.Env = append(os.Environ(), "GOGC=off")
	err := icmd.Run()
	result := []string{}
	if len(buf.String()) < 2 {
		LogError("Error fetching project dependencies: Bad response")
		Debugf("Errors: %+v", errorB.String())
		return result, nil
	}
	depsStr := strings.Replace(buf.String(), " ", "\",\"", -1)
	depsStr = strings.Replace(depsStr, "]\n[", "\",\"", -1)
	depsStr = "[\"" + depsStr[1:len(depsStr)-2] + "\"]"
	var dependencies = []string{}
	err = json.Unmarshal([]byte(depsStr), &dependencies)
	if err != nil {
		LogError("Error fetching project dependencies: %v", err)
		return nil, err
	}
	var handledDependencies = make(map[string]bool)
	for _, d := range dependencies {
		if !handledDependencies[d] {
			result = append(result, d)
			handledDependencies[d] = true
		}
	}
	return dependencies, nil
}

func InstallGoDependencies(names ...string) error {
	if len(names) < 0 {
		return nil
	}
	args := []string{"get"}
	args = append(args, names...)
	icmd := exec.Command("go", args...)
	icmd.Env = append(os.Environ(), "GOGC=off")
	return icmd.Run()

}
