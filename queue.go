package bizlogger

import (
	"encoding/json"
	"log/slog"
	"time"
)

// processQueue runs as a background worker to periodically process logs.
//
// It gets processQueueTime parameter which shows the duration between processing intervals.
// It runs in dedicated goroutine started during Logger initialization
// and immediately processes remaining logs and cleans up resources on shutdown signal.
func (l *Logger) processQueue(processQueueTime time.Duration) {
	defer l.wg.Done()
	ticker := time.NewTicker(processQueueTime)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			l.processValuesFromQueue()
		case <-l.stopChan:
			l.processValuesFromQueue()
			return
		}
	}
}

// processValuesFromFile reads and processes logs from queue
// by working with each queued log entry and  attempts file insertion for each record.
// Queue will be cleaned after successful processing.
// If any internal error occures it will be logged by techLogger to output.
func (l *Logger) processValuesFromQueue() {
	if len(l.queue) == 0 {
		return
	}
	for i := range l.queue {
		bizlog := l.queue[i]
		bizlogData, err := json.Marshal(bizlog)
		if err != nil {
			l.techLogger.Error("processValues json.Marshal", slog.Any("error", err))
			return
		}

		if _, err := l.file.Write(bizlogData); err != nil {
			l.techLogger.Error("processValues l.file.Write", slog.Any("error", err))
			return
		}
	}
	l.queue = nil

}
