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
}

// Create a new console logger
func NewConsoleLogger(o int) (l *ConsoleLogger) {
	l = &ConsoleLogger{os.Stdout, o, TIMEFORMAT}
	return
}

// Generic output function. Outputs messages if they are higher level than outLevel for this
// specific logger. If msg does not end with a newline, one will be appended.
func (l *ConsoleLogger) Output(msg *Message) {
	if msg.level < l.outLevel {
		return
	}

	buf := []byte{}
	buf = append(buf, msg.time.Format(l.timeFormat)...)
	buf = append(buf, ' ')
	buf = append(buf, levels[msg.level]...)
	buf = append(buf, ' ')
	buf = append(buf, msg.m...)
	if len(msg.m) > 0 && msg.m[len(msg.m)-1] != '\n' {
		buf = append(buf, '\n')
	}
	l.out.Write(buf)
}

func (l *ConsoleLogger) TimeFormat(f string) {
	l.timeFormat = f
}

func (l *ConsoleLogger) Fatal(format string, v ...interface{}) {
	l.Output(&Message{FATAL, fmt.Sprintf(format, v...), time.Now()})
}

func (l *ConsoleLogger) Error(format string, v ...interface{}) {
	l.Output(&Message{ERROR, fmt.Sprintf(format, v...), time.Now()})
}

func (l *ConsoleLogger) Warn(format string, v ...interface{}) {
	l.Output(&Message{WARN, fmt.Sprintf(format, v...), time.Now()})
}

func (l *ConsoleLogger) Info(format string, v ...interface{}) {
	l.Output(&Message{INFO, fmt.Sprintf(format, v...), time.Now()})
}

func (l *ConsoleLogger) Debug(format string, v ...interface{}) {
	l.Output(&Message{DEBUG, fmt.Sprintf(format, v...), time.Now()})
}

func (l *ConsoleLogger) Trace(format string, v ...interface{}) {
	l.Output(&Message{TRACE, fmt.Sprintf(format, v...), time.Now()})
}
