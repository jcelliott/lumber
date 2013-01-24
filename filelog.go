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
	TRUNC  = -1
	APPEND = 0
)

type FileLogger struct {
	queue      chan Message
	out        *os.File
	outLevel   int
	timeFormat string
}

func NewFileLogger(f string, o, mode int) (l *FileLogger, err error) {
	var file *os.File
	if mode == TRUNC {
		file, err = os.OpenFile(f, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	} else if mode == APPEND {
		file, err = os.OpenFile(f, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	} else if mode > 0 {
		file, err = openBackup(f, mode)
	} else {
		err = fmt.Errorf("Invalid mode parameter: %d", mode)
		return
	}
	if err != nil {
		err = fmt.Errorf("Error opening file '%s' for logging: %s", f, err)
		return
	}

	l = &FileLogger{make(chan Message, MSGBUFSIZE), file, o, TIMEFORMAT}

	return
}

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

func backup(f *os.File) error {
	return os.Rename(f.Name(), fmt.Sprintf("%s.1", f.Name()))
}

func (l *FileLogger) Output(msg *Message) {
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

func (l *FileLogger) TimeFormat(f string) {
	l.timeFormat = f
}

func (l *FileLogger) Level(o int) {
	if o >= TRACE && o <= FATAL {
		l.outLevel = o
	}
}

func (l *FileLogger) MsgBufSize(s int) {
	if s >= 0 {
		l.queue = make(chan Message, MSGBUFSIZE)
	}
}

func (l *FileLogger) Close() (err error) {
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

func (l *FileLogger) Fatal(format string, v ...interface{}) {
	l.Output(&Message{FATAL, fmt.Sprintf(format, v...), time.Now()})
}

func (l *FileLogger) Error(format string, v ...interface{}) {
	l.Output(&Message{ERROR, fmt.Sprintf(format, v...), time.Now()})
}

func (l *FileLogger) Warn(format string, v ...interface{}) {
	l.Output(&Message{WARN, fmt.Sprintf(format, v...), time.Now()})
}

func (l *FileLogger) Info(format string, v ...interface{}) {
	l.Output(&Message{INFO, fmt.Sprintf(format, v...), time.Now()})
}

func (l *FileLogger) Debug(format string, v ...interface{}) {
	l.Output(&Message{DEBUG, fmt.Sprintf(format, v...), time.Now()})
}

func (l *FileLogger) Trace(format string, v ...interface{}) {
	l.Output(&Message{TRACE, fmt.Sprintf(format, v...), time.Now()})
}
