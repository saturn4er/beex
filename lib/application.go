package lib

import (
	"os"

	"errors"

	"path/filepath"

	"os/exec"
	"runtime"

	"time"

	"sync"

	"github.com/howeyc/fsnotify"
)

type Application struct {
	appname        string
	path           string
	ExecutablePath string
	Args           []string
	PathsToWatch   []string
	process        *exec.Cmd
	GOPATH         string
	StopC          chan int
	RunLock        sync.Mutex
	BuildOnce      sync.Once
}

func (a *Application) GetAppname() string {
	return a.appname
}
func (a *Application) BuildAndRun() error {
	err := a.Build()
	if err != nil {
		return err
	}
	a.Run()
	return nil
}
func (a *Application) Build() error {
	var err error
	a.BuildOnce.Do(func() {
		err = a.InstallDependencies()
		if err != nil {
			return
		}
		LogInfo("Building to ./%s", a.ExecutablePath)

		args := []string{"build"}
		args = append(args, "-o", a.ExecutablePath)
		bcmd := exec.Command("go", args...)
		bcmd.Env = append(os.Environ(), "GOGC=off")
		bcmd.Stdout = os.Stdout
		bcmd.Stderr = os.Stderr
		err = bcmd.Run()
		if err != nil {
			return
		}
		if !bcmd.ProcessState.Success() {
			err = errors.New("Building error")
		}
	})
	a.BuildOnce = sync.Once{}
	return err
}
func (a *Application) Run() {
	a.RunLock.Lock()
	defer a.RunLock.Unlock()

	Debugf(a.ExecutablePath)
	a.StopC = make(chan int)
	cmd := exec.Command(a.ExecutablePath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Args = append([]string{a.ExecutablePath}, a.Args...)
	cmd.Env = os.Environ() // TODO: Add from config
	a.process = cmd
	go cmd.Run()
}
func (a *Application) Stop() error {
	a.RunLock.Lock()
	defer a.RunLock.Unlock()
	if a.process != nil {
		if a.process.ProcessState != nil {
			Debugf("Already stopped")
			return errors.New("Already stopped")
		}
		LogInfo("Stopping application..")
		return a.process.Process.Kill()
	}
	return errors.New("Not running yet")

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

	// Update file watcher
	NewFoldersWatcher(&WatchConfig{
		PathsToWath: a.PathsToWatch,
		Extensions:  []string{".go", ".tpl"},
		StopC:       a.StopC,
		Callback: func(e *fsnotify.FileEvent) error {
			// Skip TMP files for Sublime Text.
			if checkTMPFile(e.Name) {
				return nil
			}
			err := a.Stop()
			if err != nil {
				LogInfo("Failed to stop application: %v", err)
				return err
			}
			LogInfo("Successfully stopped.")
			return a.BuildAndRun()
		},
	})
	// Program stopped watcher
	go func() {
		for {
			if a.process != nil && a.process.ProcessState != nil && a.process.ProcessState.Exited() {
				if !a.process.ProcessState.Success() {
					LogInfo("Application stopped with exit code != 0.")
					close(a.StopC)
					break
				}
				LogInfo("Application stopped. Restarting...")
				a.Run()
			}
			time.Sleep(time.Second)
		}
	}()
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
func NewApplication(appname, path string, args []string) (*Application, error) {
	result := new(Application)
	result.appname = appname
	result.path = path
	result.Args = args // TODO: Add from config
	// resolve project subdirectories
	subPaths, err := getSubDirectories(path)
	if err != nil {
		return nil, err
	}
	result.PathsToWatch = subPaths // TODO: Add from config

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
	result.ExecutablePath = o // TODO: Add from config
	return result, nil
}
