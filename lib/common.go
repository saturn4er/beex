package lib

import (
	"os"
	"strings"
	"time"

	"github.com/howeyc/fsnotify"
)

type WatchConfig struct {
	PathsToWath []string
	Extensions  []string
	Callback    func(*fsnotify.FileEvent) error
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

func NewFoldersWatcher(c *WatchConfig) chan int {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		ColorLog("[ERRO] Fail to create new Watcher[ %s ]\n", err)
		os.Exit(2)
	}
	eventTime := make(map[string]int64)
	stopC := make(chan int)
	go func() {
		var lastUpdateTime time.Time
		for {
			select {
			case _, ok := <-stopC:
				if ok {
					close(stopC)
					break
				}
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
						close(stopC)
					}
					return
				}(lastUpdateTime)
			case err := <-watcher.Error:
				LogWarning("%s", err.Error()) // No need to exit here
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
	return stopC
}
