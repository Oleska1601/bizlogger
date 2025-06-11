package bizlogger

import (
	"encoding/json"
	"log/slog"
	"time"
)

func (l *Logger) processQueue() {
	defer l.wg.Done()
	ticker := time.NewTicker(time.Minute * 10) //??? какое лучше время
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			l.processValuesFromQueue()
		case <-l.stopChan:
			// пытаемся записать оставшиеся логи + выход
			l.processValuesFromQueue()
			return
		}
	}
}

func (l *Logger) processValuesFromQueue() {
	if len(l.queue) == 0 {
		return
	}
	for i := 0; i < len(l.queue); i++ {
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
