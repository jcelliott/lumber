package lumber

import (
	"fmt"
	"io"
	"os"
	"time"
)

type ConsoleLogger struct {
	out        io.WriteCloser
	outLevel   int
	timeFormat string
	prefix     string
	levels     []string
}

// Create a new console logger with output level o, and an empty prefix
func NewConsoleLogger(o int) *ConsoleLogger {
	return &ConsoleLogger{os.Stdout, o, TIMEFORMAT, "", levels}
}

func NewBasicLogger(f io.WriteCloser, level int) *ConsoleLogger {
	return &ConsoleLogger{f, level, TIMEFORMAT, "", levels}
}

// Generic output function. Outputs messages if they are higher level than outLevel for this
// specific logger. If msg does not end with a newline, one will be appended.
func (l *ConsoleLogger) output(msg *Message) {
	if msg.level < l.outLevel {
		return
	}

	buf := []byte{}
	buf = append(buf, msg.time.Format(l.timeFormat)...)
	if l.prefix != "" {
		buf = append(buf, ' ')
		buf = append(buf, l.prefix...)
	}
	buf = append(buf, ' ')
	buf = append(buf, l.levels[msg.level]...)
	buf = append(buf, ' ')
	buf = append(buf, msg.m...)
	if len(msg.m) > 0 && msg.m[len(msg.m)-1] != '\n' {
		buf = append(buf, '\n')
	}
	l.out.Write(buf)
}

// Sets the available levels for this logger
func (l *ConsoleLogger) SetLevels(lvls []string) {
	if lvls[len(lvls)-1] != "*LOG*" {
		lvls = append(lvls, "*LOG*")
	}
	l.levels = lvls
}

// Sets the output level for this logger
func (l *ConsoleLogger) Level(o int) {
	if o >= 0 && o <= len(l.levels)-1 {
		l.outLevel = o
	}
}

// Sets the prefix for this logger
func (l *ConsoleLogger) Prefix(p string) {
	l.prefix = p
}

// Sets the time format for this logger
func (l *ConsoleLogger) TimeFormat(f string) {
	l.timeFormat = f
}

// Close the logger
func (l *ConsoleLogger) Close() {
	l.output(&Message{len(l.levels) - 1, "Closing log now", time.Now()})
	l.out.Close()
}

// Logging functions
func (l *ConsoleLogger) Fatal(format string, v ...interface{}) {
	l.output(&Message{FATAL, fmt.Sprintf(format, v...), time.Now()})
}

func (l *ConsoleLogger) Error(format string, v ...interface{}) {
	l.output(&Message{ERROR, fmt.Sprintf(format, v...), time.Now()})
}

func (l *ConsoleLogger) Warn(format string, v ...interface{}) {
	l.output(&Message{WARN, fmt.Sprintf(format, v...), time.Now()})
}

func (l *ConsoleLogger) Info(format string, v ...interface{}) {
	l.output(&Message{INFO, fmt.Sprintf(format, v...), time.Now()})
}

func (l *ConsoleLogger) Debug(format string, v ...interface{}) {
	l.output(&Message{DEBUG, fmt.Sprintf(format, v...), time.Now()})
}

func (l *ConsoleLogger) Trace(format string, v ...interface{}) {
	l.output(&Message{TRACE, fmt.Sprintf(format, v...), time.Now()})
}

func (l *ConsoleLogger) Print(lvl int, v ...interface{}) {
	l.output(&Message{lvl, fmt.Sprint(v...), time.Now()})
}

func (l *ConsoleLogger) Printf(lvl int, format string, v ...interface{}) {
	l.output(&Message{lvl, fmt.Sprintf(format, v...), time.Now()})
}

func (l *ConsoleLogger) GetLevel() int {
	return l.outLevel
}

func (l *ConsoleLogger) IsFatal() bool {
	return l.outLevel <= FATAL
}

func (l *ConsoleLogger) IsError() bool {
	return l.outLevel <= ERROR
}

func (l *ConsoleLogger) IsWarn() bool {
	return l.outLevel <= WARN
}

func (l *ConsoleLogger) IsInfo() bool {
	return l.outLevel <= INFO
}

func (l *ConsoleLogger) IsDebug() bool {
	return l.outLevel <= DEBUG
}

func (l *ConsoleLogger) IsTrace() bool {
	return l.outLevel <= TRACE
}
