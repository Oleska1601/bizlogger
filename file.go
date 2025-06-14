package bizlogger

import (
	"bufio"
	"encoding/json"
	"log/slog"
	"time"
)

// insertLogInFile appends a business log record to the log file
// which is used as secondary storage when DB insertion fails.
//
// It gets bizlogData as JSON-encoded log entry bytes
// and appends newline after each entry.
// The function returns an error if file writing failed.
func (l *Logger) insertLogInFile(bizlogData []byte) error {
	_, err := l.file.Write(append(bizlogData, '\n'))
	if err != nil {
		return err
	}
	return nil
}

// processFile runs as a background worker to periodically process logs.
//
// It gets processFileTime parameter which shows the duration between processing intervals.
// It runs in dedicated goroutine started during Logger initialization
// and immediately processes remaining logs and cleans up resources on shutdown signal.
func (l *Logger) processFile(processFileTime time.Duration) {
	defer l.wg.Done()
	ticker := time.NewTicker(processFileTime)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			l.processValuesFromFile()
		case <-l.stopChan:
			l.processValuesFromFile()
			return
		}
	}
}

// processValuesFromFile reads and processes logs from fallback file
// by scanning log file line by line and attempts DB insertion for each record.
// If any internal error occures it will be logged by techLogger to output.
func (l *Logger) processValuesFromFile() {
	scanner := bufio.NewScanner(l.file)
	for scanner.Scan() {
		line := scanner.Text()
		var bizlog BizLog
		if err := json.Unmarshal([]byte(line), &bizlog); err != nil {
			l.techLogger.Error("processValuesFromFile", slog.Any("error", err))
			return
		}
		if err := l.repo.InsertLogInDB(l.ctx, &bizlog); err != nil {
			l.techLogger.Error("processValuesFromFile", slog.Any("error", err))
			return
		}

	}

	if err := scanner.Err(); err != nil {
		l.techLogger.Error("processValuesFromFile", slog.Any("error", err))
		return
	}
	l.queue = []BizLog{}
}
