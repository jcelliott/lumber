package lumber

import (
	"fmt"
	"io"
	"os"
	"time"
)

type ConsoleLogger struct {
	out        io.Writer
	outLevel   int
	timeFormat string
	prefix     string
}

// Create a new console logger with output level o, and an empty prefix
func NewConsoleLogger(o int) (l *ConsoleLogger) {
	l = &ConsoleLogger{os.Stdout, o, TIMEFORMAT, ""}
	return
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
	buf = append(buf, levels[msg.level]...)
	buf = append(buf, ' ')
	buf = append(buf, msg.m...)
	if len(msg.m) > 0 && msg.m[len(msg.m)-1] != '\n' {
		buf = append(buf, '\n')
	}
	l.out.Write(buf)
}

// Sets the output level for this logger
func (l *ConsoleLogger) Level(o int) {
	if o >= TRACE && o <= FATAL {
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

// Close the logger (For a console logger this is just a noop)
func (l *ConsoleLogger) Close() error {
	return nil
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
