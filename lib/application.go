package lib

import (
	"os"

	"errors"

	"path/filepath"

	"os/exec"
	"runtime"

	"time"

	"sync"

	"strings"

	"github.com/howeyc/fsnotify"
)

type Application struct {
	AppName        string
	SourcesPath    string
	ExecutablePath string
	Args           []string
	PathsToWatch   []string
	StopC          chan int

	// internal
	process   *exec.Cmd
	gopath    string
	runLock   sync.Mutex
	buildOnce sync.Once
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
	a.buildOnce.Do(func() {
		err = a.InstallDependencies()
		if err != nil {
			return
		}
		LogInfo("Building to %s", a.ExecutablePath)

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
	a.buildOnce = sync.Once{}
	return err
}
func (a *Application) Run() {
	a.runLock.Lock()
	defer a.runLock.Unlock()

	a.StopC = make(chan int)
	app := shouldGetRelPath(a.ExecutablePath)
	cmd := exec.Command(app)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Args = a.Args
	LogInfo("Running %v %v", app, strings.Join(cmd.Args, " "))
	cmd.Env = append(os.Environ(), Config.Envs...)
	a.process = cmd
	go cmd.Run()
}
func (a *Application) Stop() error {
	a.runLock.Lock()
	defer a.runLock.Unlock()
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
	deps, err := GetPathDeps(shouldGetRelPath(a.SourcesPath))
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
		Extensions:  append([]string{".go", ".tpl"}, Config.WatchFiles.Extensions...),
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
				if !a.process.ProcessState.Success() && !Config.WatchProcess.TryRestartOnExit {
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
func NewApplication(appname, path string, args []string) (*Application, error) {
	result := new(Application)
	result.AppName = appname
	result.SourcesPath = path
	result.Args = append(args, Config.CmdArgs...)
	// resolve project subdirectories
	subPaths, err := getSubDirectories(path)
	if err != nil {
		return nil, err
	}
	result.PathsToWatch = append(subPaths, Config.WatchFiles.Folders...)

	// Receiving GOPATH
	gps := GetGOPATHs()
	if len(gps) == 0 {
		LogError("Fail to start[ %s ]\n", "$GOPATH is not set or empty")
		return nil, errors.New("GOPATH is empty")
	}
	if len(gps) == 0 {
		return nil, errors.New("GOPATH is empty")
	}
	result.gopath = gps[0]

	// Set output path
	if Config.Build != "" {
		result.ExecutablePath = Config.Build
	} else {
		o := "./" + result.AppName
		if runtime.GOOS == "windows" {
			o += ".exe"
		}
		result.ExecutablePath, err = filepath.Abs(o)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}
