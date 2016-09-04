package lib

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	Gray = uint8(iota + 90)
	Red
	Green
	Yellow
	Blue
	Magenta
	//NRed      = uint8(31) // Normal
	EndColor = "\033[0m"

	INFO = "INFO"
	TRAC = "TRAC"
	ERRO = "ERRO"
	WARN = "WARN"
	SUCC = "SUCC"
)

var debug bool

func init() {

	d, err := strconv.ParseBool(os.Getenv("DEBUG"))
	if err != nil {
		debug = false
	}
	debug = d
}

// askForConfirmation uses Scanln to parse user input. A user must type in "yes" or "no" and
// then press enter. It has fuzzy matching, so "y", "Y", "yes", "YES", and "Yes" all count as
// confirmations. If the input is not recognized, it will ask again. The function does not return
// until it gets a valid response from the user. Typically, you should use fmt to print out a question
// before calling askForConfirmation. E.g. fmt.Println("WARNING: Are you sure? (yes/no)")
func AskForConfirmation() bool {
	var response string
	_, err := fmt.Scanln(&response)
	if err != nil {
		log.Fatal(err)
	}
	okayResponses := []string{"y", "Y", "yes", "Yes", "YES"}
	nokayResponses := []string{"n", "N", "no", "No", "NO"}
	if containsString(okayResponses, response) {
		return true
	} else if containsString(nokayResponses, response) {
		return false
	} else {
		fmt.Println("Please type yes or no and then press enter:")
		return AskForConfirmation()
	}
}
func containsString(slice []string, element string) bool {
	for _, elem := range slice {
		if elem == element {
			return true
		}
	}
	return false
}

func LogInfo(format string, a ...interface{}) {
	fmt.Println(ColorLogS("["+INFO+"] "+format, a...))
}
func LogTrace(format string, a ...interface{}) {
	fmt.Println(ColorLogS("["+TRAC+"] "+format, a...))
}
func LogError(format string, a ...interface{}) {
	fmt.Println(ColorLogS("["+ERRO+"] "+format, a...))
}
func LogWarning(format string, a ...interface{}) {
	fmt.Println(ColorLogS("["+WARN+"] "+format, a...))
}
func LogSuccess(format string, a ...interface{}) {
	fmt.Println(ColorLogS("["+SUCC+"] "+format, a...))
}

// ColorLog colors log and print to stdout.
// See color rules in function 'ColorLogS'.
func ColorLog(format string, a ...interface{}) {
	fmt.Println(ColorLogS(format, a...))
}

// ColorLogS colors log and return colored content.
// Log format: <level> <content [highlight][path]> [ error ].
// Level: TRAC -> blue; ERRO -> red; WARN -> Magenta; SUCC -> green; others -> default.
// Content: default; path: yellow; error -> red.
// Level has to be surrounded by "[" and "]".
// Highlights have to be surrounded by "# " and " #"(space), "#" will be deleted.
// Paths have to be surrounded by "( " and " )"(space).
// Errors have to be surrounded by "[ " and " ]"(space).
// Note: it hasn't support windows yet, contribute is welcome.
func ColorLogS(format string, a ...interface{}) string {
	logStr := fmt.Sprintf(format, a...)

	var clog string

	if runtime.GOOS != "windows" {
		// Level.
		i := strings.Index(logStr, "]")
		if logStr[0] == '[' && i > -1 {
			clog += "[" + getColorLevel(logStr[1:i]) + "]"
		}

		logStr = logStr[i+1:]

		// Error.
		logStr = strings.Replace(logStr, "[ ", fmt.Sprintf("[\033[%dm", Red), -1)
		logStr = strings.Replace(logStr, " ]", EndColor+"]", -1)

		// Path.
		logStr = strings.Replace(logStr, "( ", fmt.Sprintf("(\033[%dm", Yellow), -1)
		logStr = strings.Replace(logStr, " )", EndColor+")", -1)

		// Highlights.
		logStr = strings.Replace(logStr, "# ", fmt.Sprintf("\033[%dm", Gray), -1)
		logStr = strings.Replace(logStr, " #", EndColor, -1)

		logStr = clog + logStr

	} else {
		// Level.
		i := strings.Index(logStr, "]")
		if logStr[0] == '[' && i > -1 {
			clog += "[" + logStr[1:i] + "]"
		}

		logStr = logStr[i+1:]

		logStr = clog + logStr
	}

	return time.Now().Format("2006/01/02 15:04:05 ") + logStr
}

// getColorLevel returns colored level string by given level.
func getColorLevel(level string) string {
	level = strings.ToUpper(level)
	switch level {
	case INFO:
		return fmt.Sprintf("\033[%dm%s\033[0m", Blue, level)
	case TRAC:
		return fmt.Sprintf("\033[%dm%s\033[0m", Blue, level)
	case ERRO:
		return fmt.Sprintf("\033[%dm%s\033[0m", Red, level)
	case WARN:
		return fmt.Sprintf("\033[%dm%s\033[0m", Magenta, level)
	case SUCC:
		return fmt.Sprintf("\033[%dm%s\033[0m", Green, level)
	default:
		return level
	}
}

// if os.env DEBUG set, debug is on
func Debugf(format string, a ...interface{}) {
	if debug {
		_, file, line, ok := runtime.Caller(1)
		if !ok {
			file = "<unknown>"
			line = -1
		} else {
			file = filepath.Base(file)
		}
		fmt.Fprintf(os.Stderr, ColorLogS("[debug] %s:%d %s\n", file, line, format), a...)
	}
}
