package bizlogger

import (
	"encoding/json"
	"log/slog"
	"time"
)

func (l *Logger) LogUpdate(entity Entity, username string, userRole UserRole, context *string, entityID string, oldValue interface{}, newValue interface{}) {
	err := validateContextRules(userRole, context)
	bizlog := BizLog{
		Timestamp: time.Now().Format("2006-01-02T15:04:05.000Z"),
		EventType: EventTypeUpdate,
		Entity:    entity,
		Username:  username,
		UserRole:  userRole,
		Context:   context,
		EntityID:  entityID,
		OldValue:  oldValue,
		NewValue:  newValue,
	}
	bizlogData, err := json.Marshal(bizlog)
	if err != nil {
		l.techLogger.Error("LogUpdate json.Marshal", slog.Any("error", err))
		return
	}
	l.outputLogger.Println(string(bizlogData))
	err = l.repo.insertLogInDB(l.ctx, &bizlog)
	if err != nil {
		l.techLogger.Warn("LogDelete l.db.insertLogInDB", slog.Any("warn", "failed to write in DB, try write to file"))
		err = l.insertLogInFile(bizlogData)
		if err != nil {
			l.techLogger.Warn("LogDelete l.insertLogInFile", slog.Any("warn", "failed to write in file, write to queue"))
		} else {
			l.techLogger.Info("LogDelete", slog.Any("info", "log was written to file successfully"))
			return
		}
		l.queue = append(l.queue, bizlog)
		l.techLogger.Info("LogDelete", slog.Any("info", "log was written to queue successfully"))
		return
	}
}
