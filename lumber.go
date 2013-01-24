// TODO:
//  - Add an optional prefix to log messages

package lumber

import (
	"fmt"
	"time"
)

const (
	TRACE = iota
	DEBUG
	INFO
	WARN
	ERROR
	FATAL

	TIMEFORMAT = "2006/01/02 15:04:05"
	MSGBUFSIZE = 50
)

var (
	stdLog     = NewConsoleLogger(INFO)
	levels     = [...]string{"TRACE", "DEBUG", "INFO ", "WARN ", "ERROR", "FATAL"}
	timeFormat = "2006/01/02 15:04:05"
	msgBufSize = 50
)

type Logger interface {
	Fatal(format string, v ...interface{})
	Error(format string, v ...interface{})
	Warn(format string, v ...interface{})
	Info(format string, v ...interface{})
	Debug(format string, v ...interface{})
	Trace(format string, v ...interface{})
	output(level int, format string, v ...interface{})
	Close()
}

type Message struct {
	level int
	m     string
	time  time.Time
}

func TimeFormat(f string) {
	stdLog.TimeFormat(f)
}

func LvlStr(l int) string {
	if l >= TRACE && l <= FATAL {
		return levels[l]
	}
	return ""
}

func Level(o int) {
	stdLog.Level(o)
}

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
