/*
Package lumber implements a simple logger that supports log levels and rotation.
*/
package lumber

import (
	"fmt"
	"strings"
	"time"
)

const (
	TRACE = iota
	DEBUG
	INFO
	WARN
	ERROR
	FATAL
	LOG

	TIMEFORMAT = "2006/01/02 15:04:05"
)

var (
	stdLog     = NewConsoleLogger(INFO)
	levels     = [...]string{"TRACE", "DEBUG", "INFO ", "WARN ", "ERROR", "FATAL", "*LOG*"}
	timeFormat = TIMEFORMAT
)

type Logger interface {
	Fatal(string, ...interface{})
	Error(string, ...interface{})
	Warn(string, ...interface{})
	Info(string, ...interface{})
	Debug(string, ...interface{})
	Trace(string, ...interface{})
	Level(int)
	Prefix(string)
	TimeFormat(string)
	Close()
	output(msg *Message)
}

type Message struct {
	level int
	m     string
	time  time.Time
}

// Returns the string representation of the level
func LvlStr(l int) string {
	if l >= TRACE && l <= FATAL {
		return levels[l]
	}
	return ""
}

// Returns the int value of the level
func LvlInt(s string) int {
	for i, str := range levels {
		if strings.TrimSpace(str) == strings.ToUpper(s) {
			return i
		}
	}
	return 0
}

// Sets the output level for the default logger
func Level(o int) {
	stdLog.Level(o)
}

// Sets the time format for the default logger
func TimeFormat(f string) {
	stdLog.TimeFormat(f)
}

// Logging functions
func Fatal(format string, v ...interface{}) {
	stdLog.output(&Message{FATAL, fmt.Sprintf(format, v...), time.Now()})
}

func Error(format string, v ...interface{}) {
	stdLog.output(&Message{ERROR, fmt.Sprintf(format, v...), time.Now()})
}

func Warn(format string, v ...interface{}) {
	stdLog.output(&Message{WARN, fmt.Sprintf(format, v...), time.Now()})
}

func Info(format string, v ...interface{}) {
	stdLog.output(&Message{INFO, fmt.Sprintf(format, v...), time.Now()})
}

func Debug(format string, v ...interface{}) {
	stdLog.output(&Message{DEBUG, fmt.Sprintf(format, v...), time.Now()})
}

func Trace(format string, v ...interface{}) {
	stdLog.output(&Message{TRACE, fmt.Sprintf(format, v...), time.Now()})
}
