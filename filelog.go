/*
 * TODO:
 *  - Logging to a FileLogger will write the message to a channel of Messages.
 *    A separate goroutine will consume messages from the channel and write
 *    them to the file
 */

package lumber

import (
	"fmt"
	"os"
	"time"
)

const (
	// mode constants
	BACKUP = -2
	TRUNC  = -1
	APPEND = 0
)

type FileLogger struct {
	queue      chan *Message
	out        *os.File
	outLevel   int
	timeFormat string
	prefix     string
	done       chan bool
}

// Creates a new FileLogger with filename f, output level o, mode, and an empty prefix
// Modes are APPEND (append to existing log if it exists), TRUNC (truncate old log file to create
// the new one), BACKUP (moves old log to log.name.1 before creaing new log).
// buf specifies the buffer size. Set to 0 for an unbuffered logger. If a buffer is used, the log
// *must* be Close'd to prevent data loss.
func NewFileLogger(f string, o, mode, buf int) (l *FileLogger, err error) {
	var file *os.File
	switch {
	case mode == TRUNC:
		file, err = os.OpenFile(f, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	case mode == APPEND:
		file, err = os.OpenFile(f, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	case mode > 0:
		file, err = openBackup(f, mode)
	default:
		err = fmt.Errorf("Invalid mode parameter: %d", mode)
		return
	}
	if err != nil {
		err = fmt.Errorf("Error opening file '%s' for logging: %s", f, err)
		return
	}

	l = &FileLogger{make(chan *Message, buf), file, o, TIMEFORMAT, "", make(chan bool)}

	go func() {
		for {
			m, ok := <-l.queue
			if !ok {
				l.done <- true
				return
			}
			l.output(m)
		}
	}()
	return
}

// Attempt to create new log. If the file already exists, backup the old one and create a new file
func openBackup(f string, mode int) (*os.File, error) {
	// First try to open the file with O_EXCL (file must not already exist)
	file, err := os.OpenFile(f, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0666)
	if err == nil {
		return file, nil
	}
	if !os.IsExist(err) {
		return nil, fmt.Errorf("Error opening file for logging: %s", err)
	}

	// The file already exists, we need to back it up
	err = os.Rename(f, fmt.Sprintf("%s.1", f))
	if err != nil {
		backupErr := fmt.Errorf("Error backing up log: %s", err)
		file, err = os.OpenFile(f, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			return nil, fmt.Errorf("%s. Error appending to existing log file: %s", backupErr, err)
		}
		return file, backupErr
	}

	// Open new file for log
	file, err = os.OpenFile(f, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return nil, fmt.Errorf("Error opening file for logging: %s", err)
	}

	return file, err
}

// Rename "log.name" to "log.name.1"
func backup(f *os.File) error {
	return os.Rename(f.Name(), fmt.Sprintf("%s.1", f.Name()))
}

// Generic output function. Outputs messages if they are higher level than outLevel for this
// specific logger. If msg does not end with a newline, one will be appended.
func (l *FileLogger) output(msg *Message) {
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
func (l *FileLogger) Level(o int) {
	if o >= TRACE && o <= FATAL {
		l.outLevel = o
	}
}

// Sets the prefix for this logger
func (l *FileLogger) Prefix(p string) {
	l.prefix = p
}

// Sets the time format for this logger
func (l *FileLogger) TimeFormat(f string) {
	l.timeFormat = f
}

// Flush anything that hasn't been written and close the logger. If the FileLogger was created with
// a buffer size > 0, Close *must* be called to prevent losing data.
func (l *FileLogger) Close() (err error) {
	close(l.queue)
	<-l.done
	err = l.out.Sync()
	if err != nil {
		l.Error("Could not sync log file")
		err = fmt.Errorf("Could not sync log file: %s", err)
	}
	err = l.out.Close()
	if err != nil {
		l.Error("Could not close log file")
		err = fmt.Errorf("Could not close log file: %s", err)
	}
	return
}

// Logging functions
func (l *FileLogger) Fatal(format string, v ...interface{}) {
	l.queue <- &Message{FATAL, fmt.Sprintf(format, v...), time.Now()}
}

func (l *FileLogger) Error(format string, v ...interface{}) {
	l.queue <- &Message{ERROR, fmt.Sprintf(format, v...), time.Now()}
}

func (l *FileLogger) Warn(format string, v ...interface{}) {
	l.queue <- &Message{WARN, fmt.Sprintf(format, v...), time.Now()}
}

func (l *FileLogger) Info(format string, v ...interface{}) {
	l.queue <- &Message{INFO, fmt.Sprintf(format, v...), time.Now()}
}

func (l *FileLogger) Debug(format string, v ...interface{}) {
	l.queue <- &Message{DEBUG, fmt.Sprintf(format, v...), time.Now()}
}

func (l *FileLogger) Trace(format string, v ...interface{}) {
	l.queue <- &Message{TRACE, fmt.Sprintf(format, v...), time.Now()}
}
