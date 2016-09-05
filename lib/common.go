package lib

import (
	"os"
	"strings"
	"time"

	"path/filepath"

	"github.com/howeyc/fsnotify"
)

type WatchConfig struct {
	PathsToWath []string
	Extensions  []string
	Callback    func(*fsnotify.FileEvent) error
	StopC       chan int
}

// IsExist returns whether a file or directory exists.
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}

// checkTMPFile returns true if the event was for TMP files.
func checkTMPFile(name string) bool {
	if strings.HasSuffix(strings.ToLower(name), ".tmp") {
		return true
	}
	return false
}

// getFileModTime retuens unix timestamp of `os.File.ModTime` by given path.
func getFileModTime(path string) int64 {
	path = strings.Replace(path, "\\", "/", -1)
	f, err := os.Open(path)
	if err != nil {
		return 0
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return 0
	}

	return fi.ModTime().Unix()
}

func NewFoldersWatcher(c *WatchConfig) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		ColorLog("[ERRO] Fail to create new Watcher[ %s ]\n", err)
		os.Exit(2)
	}
	eventTime := make(map[string]int64)
	go func() {
		var lastUpdateTime time.Time
		var endWatch bool
		for {
			select {
			case <-c.StopC:
				Debugf("Stop watcher. Receive stop signal")
				endWatch = true
				break
			case e := <-watcher.Event:
				var fire bool
				for _, ext := range c.Extensions {
					if strings.HasSuffix(e.Name, ext) {
						fire = true
					}
				}
				if !fire {
					Debugf("Skipping event %v [ no tracked extension ]", e)
					continue
				}
				mt := getFileModTime(e.Name)
				if t := eventTime[e.Name]; mt == 0 && mt == t {
					continue
				}
				eventTime[e.Name] = mt
				ColorLog("[EVEN] %s", e)
				lastUpdateTime = time.Now()
				go func(curUpdateTime time.Time) {
					// Wait 1s before autobuild util there is no file change.
					time.Sleep(time.Second)
					// Check if there wasn't any update anymore
					if lastUpdateTime != curUpdateTime {
						return
					}
					err := c.Callback(e)
					if err != nil {
						Debugf("Stop watcher. Received error: %v", err.Error())
						close(c.StopC)
					}
					return
				}(lastUpdateTime)
			case err := <-watcher.Error:
				LogWarning("%s", err.Error()) // No need to exit here
			}
			if endWatch == true {
				break
			}
		}
	}()

	for _, p := range c.PathsToWath {
		Debugf("Watch directory( %s )", p)
		err = watcher.Watch(p)
		if err != nil {
			LogError("Fail to watch directory[ %s ]", err)
			os.Exit(2)
		}
	}
	LogInfo("Watching %d directories...", len(c.PathsToWath))
}

func shouldGetRelPath(dest string) string {
	wd, err := os.Getwd()
	if err != nil {
		panic("Can't get working directory: " + err.Error())
	}
	result, err := filepath.Rel(wd, dest)
	if err != nil {
		panic("Can't get relative path to application")
	}
	if len(result) > 0 && result[0] != '.' {
		result = "./" + result
	}
	return result
}
