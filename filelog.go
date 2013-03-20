package lumber

import (
	"bufio"
	"fmt"
	"os"
	// "path/filepath"
	"time"
)

const (
	// mode constants
	TRUNC  = -1
	APPEND = 0
)

type FileLogger struct {
	out        *os.File
	outLevel   int
	timeFormat string
	prefix     string
	lineMode   int
	curLines   int
	rotateMode int
}

// Creates a new FileLogger with filename f, output level o, mode, and an empty prefix
// Modes are APPEND (append to existing log if it exists), TRUNC (truncate old log file to create
// the new one), BACKUP (moves old log to log.name.1 before creaing new log).

// lineMode: -1 == truncate, 0 == no limit
// rotateModeMode: 0 == no backups

func NewFileLogger(f string, o, lineMode, rotateMode int) (l *FileLogger, err error) {
	var file *os.File

	if rotateMode <= 0 { // no rotation
		switch {
		case lineMode == TRUNC:
			// just truncate file and start logging
			file, err = os.OpenFile(f, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
		case lineMode == APPEND:
			// open log file, append if it already exists
			file, err = os.OpenFile(f, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
		case lineMode > 0:
			// line limit, but no backups: file-internal rotation
			// file, err = os.OpenFile(f, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
			err = fmt.Errorf("File-internal rotation is not implemented yet")
			return
		default:
			err = fmt.Errorf("Invalid mode parameter: %d", lineMode)
			return
		}
	} else { // rotation
		switch {
		case lineMode == TRUNC:
			// rotateMode every time a new logger is created
			file, err = openBackup(f, lineMode)
		case lineMode == APPEND:
			// this doesn't make any sense
			err = fmt.Errorf("Cannot use APPEND with log rotation")
			return
		case lineMode > 0:
			// "normal" rotation, when file reaches line limit
			file, err = openBackup(f, lineMode)
		default:
			err = fmt.Errorf("Invalid mode parameter: %d", lineMode)
			return
		}
	}

	if err != nil {
		err = fmt.Errorf("Error opening file '%s' for logging: %s", f, err)
		return
	}

	l = &FileLogger{
		out:        file,
		outLevel:   o,
		timeFormat: TIMEFORMAT,
		prefix:     "",
		lineMode:   lineMode,
		curLines:   0,
		rotateMode: rotateMode,
	}

	return
}

// Attempt to create new log. Specific behavior depends on the lineMode setting
func openBackup(f string, lineMode int) (*os.File, error) {
	// First try to open the file with O_EXCL (file must not already exist)
	file, err := os.OpenFile(f, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0644)
	// If there are no errors, we can just use this file
	if err == nil {
		return file, nil
	}
	// If the error wasn't an 'Exist' error, we've got a problem
	if !os.IsExist(err) {
		return nil, fmt.Errorf("Error opening file for logging: %s", err)
	}

	// The file already exists, open it
	file, err = os.OpenFile(f, os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("Error opening file for logging: %s", err)
	}

	// If we're considering lines, check the line count
	if lineMode > 0 {
		// If its still under the line limit, just return the file
		if countLines(file) < lineMode {
			return file, nil
		}
	}

	// Either we're always rotating or we're over the line limit, we need to rotate
	file, err = doRotate(f)
	if err != nil {
		return nil, fmt.Errorf("Error opening file for logging: %s", err)
	}

	return file, err
}

// Rotate the logs
func (l *FileLogger) rotate() error {
	file, err := doRotate(l.out.Name())
	if err != nil {
		return fmt.Errorf("Error rotating logs: %s", err)
	}
	l.out = file
	return nil
}

// Rotate all the logs and return newly vacated file
// Rename 'log.name' to 'log.name.1' and 'log.name.1' to 'log.name.2' etc
func doRotate(f string) (*os.File, error) {
	// get all rotated files
	// TODO: implement this with a real regex so we don't accidentally get a file we shouldn't
	// list, err := filepath.Glob(fmt.Sprintf("%s.*", f))
	// reverse sort files
	// for file := range list {
	// strings.LastIndex(some stuff here)
	// get the integer part
	// increment it
	// rename file with that extension
	// }
	err := os.Rename(f, fmt.Sprintf("%s.1", f))
	if err != nil {
		return nil, fmt.Errorf("Error rotating logs: %s", err)
	}
	file, err := os.OpenFile(f, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("Error opening file for logging: %s", err)
	}
	return file, nil
}

// Generic output function. Outputs messages if they are higher level than outLevel for this
// specific logger. If msg does not end with a newline, one will be appended.
func (l *FileLogger) output(msg *Message) {
	if msg.level < l.outLevel {
		return
	}
	if l.lineMode > 0 && l.curLines >= l.lineMode {
		err := l.rotate()
		if err != nil {
			Error("Error backing up log:", err)
		}
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
	l.curLines += 1
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

// Flush anything that hasn't been written and close the logger
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

// return the number of lines in the given file
func countLines(f *os.File) int {
	r := bufio.NewReader(f)
	count := 0
	var err error = nil
	for err == nil {
		prefix := true
		_, prefix, err = r.ReadLine()
		if err != nil {
		}
		// sometimes we don't get the whole line at once
		if !prefix && err == nil {
			count++
		}
	}
	return count
}

// Logging functions
func (l *FileLogger) Fatal(format string, v ...interface{}) {
	l.output(&Message{FATAL, fmt.Sprintf(format, v...), time.Now()})
}

func (l *FileLogger) Error(format string, v ...interface{}) {
	l.output(&Message{ERROR, fmt.Sprintf(format, v...), time.Now()})
}

func (l *FileLogger) Warn(format string, v ...interface{}) {
	l.output(&Message{WARN, fmt.Sprintf(format, v...), time.Now()})
}

func (l *FileLogger) Info(format string, v ...interface{}) {
	l.output(&Message{INFO, fmt.Sprintf(format, v...), time.Now()})
}

func (l *FileLogger) Debug(format string, v ...interface{}) {
	l.output(&Message{DEBUG, fmt.Sprintf(format, v...), time.Now()})
}

func (l *FileLogger) Trace(format string, v ...interface{}) {
	l.output(&Message{TRACE, fmt.Sprintf(format, v...), time.Now()})
}
