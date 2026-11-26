package lumber

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func newCloseTestLogger(t *testing.T) (*FileLogger, string, func()) {
	dir, err := ioutil.TempDir("", "lumber-close-")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "test.log")
	logger, err := NewAppendLogger(path)
	if err != nil {
		os.RemoveAll(dir)
		t.Fatal(err)
	}
	return logger, path, func() { os.RemoveAll(dir) }
}

func closeResult(closeLogger func()) <-chan interface{} {
	result := make(chan interface{}, 1)
	go func() {
		defer func() { result <- recover() }()
		closeLogger()
	}()
	return result
}

func awaitClose(t *testing.T, result <-chan interface{}) {
	select {
	case failure := <-result:
		if failure != nil {
			t.Errorf("Close panicked: %v", failure)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Close did not return")
	}
}

func checkClosedLog(t *testing.T, logger *FileLogger, path string, messages int) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Count(string(data), "queued message"); got != messages {
		t.Errorf("Close returned with %d messages written, want %d", got, messages)
	}
	if got := strings.Count(string(data), "Closing log now"); got != 1 {
		t.Errorf("got %d shutdown messages, want 1", got)
	}
	if _, err := logger.out.Stat(); err == nil {
		t.Error("Close returned before the output file was closed")
	}
}

func TestFileLoggerCloseRepeated(t *testing.T) {
	for _, messages := range []int{0, 3} {
		logger, path, cleanup := newCloseTestLogger(t)
		defer cleanup()
		for i := 0; i < messages; i++ {
			logger.Info("queued message %d", i)
		}
		for i := 0; i < 3; i++ {
			awaitClose(t, closeResult(logger.Close))
			checkClosedLog(t, logger, path, messages)
		}
	}
}

func TestFileLoggerConcurrentCloseWaitsForOutput(t *testing.T) {
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	defer writer.Close()
	logger := NewBasicFileLogger(writer, INFO)
	// More than the pipe can buffer: shutdown cannot finish until it is read.
	message := strings.Repeat("queued message\n", 1<<17)
	logger.Info("%s", message)
	const callers = 8
	results := make(chan interface{}, callers)
	start := make(chan struct{})
	for i := 0; i < callers; i++ {
		go func() {
			defer func() { results <- recover() }()
			<-start
			logger.Close()
		}()
	}
	close(start)
	completed := 0
	select {
	case failure := <-results:
		completed++
		t.Errorf("Close returned before buffered output was drained (panic: %v)", failure)
	case <-time.After(100 * time.Millisecond):
	}
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	for ; completed < callers; completed++ {
		awaitClose(t, results)
	}
	if !strings.Contains(string(data), message) {
		t.Error("Close lost queued output")
	}
	if got := strings.Count(string(data), "Closing log now"); got != 1 {
		t.Errorf("got %d shutdown messages, want 1", got)
	}
}

func TestFileLoggerCloseAfterShutdown(t *testing.T) {
	logger, path, cleanup := newCloseTestLogger(t)
	defer cleanup()
	logger.Info("queued message")
	// Rotation errors initiate shutdown without waiting for it to finish.
	logger.close()
	awaitClose(t, closeResult(logger.Close))
	checkClosedLog(t, logger, path, 1)
}

func TestMultiLoggerCloseSharedFileLogger(t *testing.T) {
	logger, path, cleanup := newCloseTestLogger(t)
	defer cleanup()
	logger.Info("queued message")
	multi := NewMultiLogger()
	multi.AddLoggers(logger, logger)
	awaitClose(t, closeResult(multi.Close))
	awaitClose(t, closeResult(multi.Close))
	checkClosedLog(t, logger, path, 1)
}
