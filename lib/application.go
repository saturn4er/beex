package lib

import (
	"os"

	"errors"

	"path/filepath"

	"os/exec"
	"runtime"

	"github.com/howeyc/fsnotify"
)

type Application struct {
	appname        string
	path           string
	ExecutablePath string
	PathsToWatch   []string
	ExitC          chan int
	GOPATH         string
	RunC           chan int
}

func (a *Application) GetAppname() string {
	return a.appname
}
func (a *Application) BuildAndRun() error {
	err := a.Build()
	if err != nil {
		return err
	}
	err = a.Run()
	if err != nil {
		return err
	}
	return nil
}
func (a *Application) Build() error {
	err := a.InstallDependencies()
	if err != nil {
		return err
	}
	LogInfo("Building to ./%s", a.ExecutablePath)

	args := []string{"build"}
	args = append(args, "-o", a.ExecutablePath)
	bcmd := exec.Command("go", args...)
	bcmd.Env = append(os.Environ(), "GOGC=off")
	bcmd.Stdout = os.Stdout
	bcmd.Stderr = os.Stderr
	err = bcmd.Run()
	return nil
}
func (a *Application) Run() error {
	return nil
}
func (a *Application) InstallDependencies() error {
	LogInfo("Checking dependencies...")
	deps, err := GetPathDeps(a.relPath())
	if err != nil {
		return err
	}
	var o int
	// installing by 20 deps (command have limit in size)
	for {
		to := o + 20
		if to > len(deps) {
			to = len(deps)
		}
		InstallGoDependencies(deps[o:to]...)
		o = to
		if o >= len(deps) {
			break
		}

	}
	LogInfo("Checking dependencies complete")
	return nil
}
func (a *Application) RunRestartWatcher() {
	NewFoldersWatcher(&WatchConfig{
		PathsToWath: a.PathsToWatch,
		Extensions:  []string{".go", ".tpl"},
		Callback: func(e *fsnotify.FileEvent) error {
			// Skip TMP files for Sublime Text.
			if checkTMPFile(e.Name) {
				return nil
			}
			return a.BuildAndRun()
		},
	})
}
func (a *Application) relPath() string {
	wd, err := os.Getwd()
	if err != nil {
		panic("Can't get working directory: " + err.Error())
	}
	result, err := filepath.Rel(wd, a.path)
	if err != nil {
		panic("Can't get relative path to application")
	}
	return result
}
func NewApplication(appname, path string) (*Application, error) {
	result := new(Application)
	result.appname = appname
	result.path = path

	// resolve project subdirectories
	subPaths, err := getSubDirectories(path)
	if err != nil {
		return nil, err
	}
	result.PathsToWatch = subPaths

	// Receiving GOPATH
	gps := GetGOPATHs()
	if len(gps) == 0 {
		LogError("Fail to start[ %s ]\n", "$GOPATH is not set or empty")
		return nil, errors.New("GOPATH is empty")
	}
	if len(gps) == 0 {
		return nil, errors.New("GOPATH is empty")
	}
	result.GOPATH = gps[0]

	// Set output path
	o := result.appname
	if runtime.GOOS == "windows" {
		o += ".exe"
	}
	result.ExecutablePath = o
	return result, nil
}
