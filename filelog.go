package lumber

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	// mode constants
	APPEND = iota
	TRUNC
	BACKUP
	ROTATE
)

type FileLogger struct {
	out        *os.File
	outLevel   int
	timeFormat string
	prefix     string
	maxLines   int
	curLines   int
	maxRotate  int
	mode       int
}

// Convenience function to create a new append-only logger
func NewAppendLogger(f string) (*FileLogger, error) {
	return NewFileLogger(f, INFO, APPEND, 0, 0)
}

// Convenience function to create a new truncating logger
func NewTruncateLogger(f string) (*FileLogger, error) {
	return NewFileLogger(f, INFO, TRUNC, 0, 0)
}

// Convenience function to create a new backup logger
func NewBackupLogger(f string, maxBackup int) (*FileLogger, error) {
	return NewFileLogger(f, INFO, BACKUP, 0, maxBackup)
}

// Convenience function to create a new rotating logger
func NewRotateLogger(f string, maxLines, maxRotate int) (*FileLogger, error) {
	return NewFileLogger(f, INFO, ROTATE, maxLines, maxRotate)
}

// Creates a new FileLogger with filename f, output level o, and an empty prefix.
// Modes are described in the documentation; maxLines and maxRotate are only significant
// for some modes.
func NewFileLogger(f string, o, mode, maxLines, maxRotate int) (l *FileLogger, err error) {
	var file *os.File

	switch mode {
	case APPEND:
		// open log file, append if it already exists
		file, err = os.OpenFile(f, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	case TRUNC:
		// just truncate file and start logging
		file, err = os.OpenFile(f, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	case BACKUP:
		// rotateMode every time a new logger is created
		file, err = openBackup(f, 0, maxRotate)
	case ROTATE:
		// "normal" rotation, when file reaches line limit
		file, err = openBackup(f, maxLines, maxRotate)
	default:
		return nil, fmt.Errorf("Invalid mode parameter: %d", mode)
	}
	if err != nil {
		return nil, fmt.Errorf("Error creating logger: %s", err)
	}

	l = &FileLogger{
		out:        file,
		outLevel:   o,
		timeFormat: TIMEFORMAT,
		prefix:     "",
		maxLines:   maxLines,
		curLines:   0,
		mode:       mode,
	}

	if mode == ROTATE {
		// get the current line count if relevant
		l.curLines = countLines(l.out)
	}
	return
}

// Attempt to create new log. Specific behavior depends on the maxLines setting
func openBackup(f string, maxLines, maxRotate int) (*os.File, error) {
	// first try to open the file with O_EXCL (file must not already exist)
	file, err := os.OpenFile(f, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0644)
	// if there are no errors (it's a new file), we can just use this file
	if err == nil {
		return file, nil
	}
	// if the error wasn't an 'Exist' error, we've got a problem
	if !os.IsExist(err) {
		return nil, fmt.Errorf("Error opening file for logging: %s", err)
	}

	if maxLines == 0 {
		// we're in backup mode, rotate and return the new file
		return doRotate(f, maxRotate)
	}

	// the file already exists, open it
	return os.OpenFile(f, os.O_RDWR|os.O_APPEND, 0644)
}

// Rotate the logs
func (l *FileLogger) rotate() error {
	l.output(&Message{LOG, "Rotating log", time.Now()})
	oldFile := l.out
	file, err := doRotate(l.out.Name(), l.maxRotate)
	if err != nil {
		return fmt.Errorf("Error rotating logs: %s", err)
	}
	l.curLines = 0
	l.out = file
	oldFile.Close()
	return nil
}

// Rotate all the logs and return a file with newly vacated filename
// Rename 'log.name' to 'log.name.1' and 'log.name.1' to 'log.name.2' etc
func doRotate(f string, limit int) (*os.File, error) {
	// create a format string with the correct amount of zero-padding for the limit
	numFmt := fmt.Sprintf(".%%0%dd", len(fmt.Sprintf("%d", limit)))
	// get all rotated files and sort them in reverse order
	list, err := filepath.Glob(fmt.Sprintf("%s.*", f))
	if err != nil {
		return nil, fmt.Errorf("Error rotating logs: %s", err)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(list)))
	for _, file := range list {
		parts := strings.Split(file, ".")
		numPart := parts[len(parts)-1]
		num, err := strconv.Atoi(numPart)
		if err != nil {
			// not a number, don't rotate it
			continue
		}
		if num >= limit {
			// we're at the limit, don't rotate it
			continue
		}
		newName := fmt.Sprintf(strings.Join(parts[:len(parts)-1], ".")+numFmt, num+1)
		if err = os.Rename(file, newName); err != nil {
			return nil, fmt.Errorf("Error rotating logs: %s", err)
		}
	}
	if err = os.Rename(f, fmt.Sprintf(f+numFmt, 1)); err != nil {
		return nil, fmt.Errorf("Error rotating logs: %s", err)
	}
	return os.OpenFile(f, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
}

// Generic output function. Outputs messages if they are higher level than outLevel for this
// specific logger. If msg does not end with a newline, one will be appended.
func (l *FileLogger) output(msg *Message) {
	if msg.level < l.outLevel {
		return
	}
	if l.mode == ROTATE && l.curLines >= l.maxLines {
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
