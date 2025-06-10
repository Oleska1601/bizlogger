package bizlogger

import (
	"bufio"
	"encoding/json"
	"log/slog"
	"time"
)

func (l *Logger) insertLogInFile(bizlogData []byte) error {
	_, err := l.file.Write(append(bizlogData, '\n'))
	if err != nil {
		return err
	}
	return nil
}

func (l *Logger) processFile() {
	defer l.wg.Done()
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			l.processValuesFromFile()
		case <-l.stopChan:
			// пытаемся записать оставшиеся логи + выход
			l.processValuesFromFile()
			return
		}
	}
}

func (l *Logger) processValuesFromFile() {
	scanner := bufio.NewScanner(l.file)
	for scanner.Scan() {
		line := scanner.Text()
		var bizlog BizLog
		if err := json.Unmarshal([]byte(line), &bizlog); err != nil {
			l.techLogger.Error("processValuesFromFile", slog.Any("error", err))
			return
		}
		if err := l.repo.insertLogInDB(l.ctx, &bizlog); err != nil {
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
