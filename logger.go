// Package bizlogger provides structured logger for different business processes.
// It records detailed information about CREATE, UPDATE, DELETE actions within the information system.
//
// Each logging action has its own structured message format.
package bizlogger

import (
	"context"
	"encoding/json"
	"log"
	"log/slog"
	"os"
	"sync"
	"time"
)

// LoggerInterface defines the methods for logger.
type LoggerInterface interface {
	// LogCreate records the creation of a new entity.
	// Only possible for userRole 'SA'.
	LogCreate(entity Entity, username string, userRole UserRole, context *string, entityID string, newValue interface{}, description *string)
	// LogDelete records the deletion of an entity.
	// Only possible for userRole 'SA'.
	LogDelete(entity Entity, username string, userRole UserRole, context *string, entityID string, description *string)
	// LogUpdate records changes to an entity.
	LogUpdate(entity Entity, username string, userRole UserRole, context *string, entityID string, oldValue interface{}, newValue interface{})
}

// pgRepoInterface defines the database methods.
type pgRepoInterface interface {
	// CreateTables creates new table by input path sql.
	// Returns error if schema creation fails.
	CreateTables(path string) error
	// InsertLogInDB get a business log record to PostgreSQL.
	// Returns error if the operation fails.
	InsertLogInDB(ctx context.Context, bizlog *BizLog) error
}

// Logger is the concrete implementation of LoggerInterface.
type Logger struct {
	ctx          context.Context // ctx is base context for lifecycle management
	repo         pgRepoInterface // repo defines methods for working with postgres database
	outputLogger *log.Logger     // outputLogger is console logger (stdout)
	file         *os.File        // file is file handle for fallback storage
	techLogger   *slog.Logger    // techLogger defines internal error logger
	wg           *sync.WaitGroup // wg defines WaitGroup for file and queue processing
	queue        []BizLog        // queue is buffer for cases when error writing to file occures
	stopChan     chan struct{}   // stopChan defines shutdown signal channel to end file and queue writing
}

// New creates a new Logger with the specified log level.
//
// ctx is base context for lifecycle management
// The repo parameter accepts interface with emthods defined in pgRepoInterface
// processFileTime and processQueueTime show intervals for file and queue processing
// (time duration depends on workload of your system specifically)
//
// function returns initialized logger, cleanup function that must be called to release resources
// and initialization error if something wrong happens.
func New(ctx context.Context, repo pgRepoInterface, processFileTime, processQueueTime time.Duration) (*Logger, func(), error) {
	techLogger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	outputLogger := log.New(os.Stdout, "", 0)
	file, err := os.Create("bizlog.json")
	if err != nil {
		techLogger.Error("bizlogger New", slog.Any("error", err))
		return nil, nil, err
	}
	err = repo.CreateTables("./query.sql")
	if err != nil {
		techLogger.Error("bizlogger New", slog.Any("error", err))
		return nil, nil, err
	}

	l := &Logger{
		ctx:          ctx,
		repo:         repo,
		outputLogger: outputLogger,
		file:         file,
		techLogger:   techLogger,
		wg:           &sync.WaitGroup{},
		queue:        []BizLog{},
		stopChan:     make(chan struct{}),
	}
	l.wg.Add(2)
	go l.processFile(processFileTime)
	go l.processQueue(processQueueTime)

	stopLogger := func() {
		close(l.stopChan)
		l.wg.Wait()
	}

	return l, stopLogger, nil
}

// LogCreate records the creation of a new entity.
// Only possible for userRole 'SA'.
//
// Arguments:
//   - entity represents the type of business entity being modified. (user/context)
//   - username of who performed the action.
//   - userRole represents the role of the user performing the action. (only SA can call LogCreate)
//   - context provides additional business context (optional).
//   - entityID identifies uniquely the affected entity.
//   - newValue contains the new state of the entity.
//   - description provides additional information or explanation of the event (optional).
func (l *Logger) LogCreate(entity Entity, username string, userRole UserRole, context *string, entityID string, newValue any, description *string) {
	if userRole != UserRoleSA {
		l.techLogger.Error("LogCreate", slog.Any("error", "user roles except SA cannot use LogCreate"))
		return
	}
	bizlog := BizLog{
		Timestamp:   time.Now().Format("2006-01-02T15:04:05.000Z"),
		EventType:   EventTypeCreate,
		Entity:      entity,
		Username:    username,
		UserRole:    userRole,
		Context:     context,
		EntityID:    entityID,
		NewValue:    newValue,
		Description: description,
	}
	bizlogData, err := json.Marshal(bizlog)
	if err != nil {
		l.techLogger.Error("LogCreate json.Marshal", slog.Any("error", err))
		return
	}
	l.outputLogger.Println(string(bizlogData))
	err = l.repo.InsertLogInDB(l.ctx, &bizlog)
	if err != nil {
		l.techLogger.Warn("LogCreate l.db.insertLogInDB", slog.Any("warn", "failed to write in DB, try write to file"))
		err = l.insertLogInFile(bizlogData)
		if err != nil {
			l.techLogger.Warn("LogCreate l.insertLogInFile", slog.Any("warn", "failed to write in file, write to queue"))
		} else {
			l.techLogger.Info("LogCreate", slog.Any("info", "log was written to file successfully"))
			return
		}
		l.queue = append(l.queue, bizlog)
		l.techLogger.Info("LogCreate", slog.Any("info", "log was written to queue successfully"))
		return
	}
}

// LogDelete records the deletion of an entity.
// Only possible for userRole 'SA'.
//
// Arguments:
//   - entity represents the type of business entity being modified. (user/context)
//   - username of who performed the action.
//   - userRole represents the role of the user performing the action. (only SA can call LogDelete)
//   - context provides additional business context (optional).
//   - entityID identifies uniquely the affected entity.
//   - description provides additional information or explanation of the event (optional).
func (l *Logger) LogDelete(entity Entity, username string, userRole UserRole, context *string, entityID string, description *string) {
	if userRole != UserRoleSA {
		l.techLogger.Error("LogCreate", slog.Any("error", "user roles except SA cannot use LogDelete"))
		return
	}
	bizlog := BizLog{
		Timestamp:   time.Now().Format("2006-01-02T15:04:05.000Z"),
		EventType:   EventTypeDelete,
		Entity:      entity,
		Username:    username,
		UserRole:    userRole,
		Context:     context,
		EntityID:    entityID,
		Description: description,
	}
	bizlogData, err := json.Marshal(bizlog)
	if err != nil {
		l.techLogger.Error("LogDelete json.Marshal", slog.Any("error", err))
		return
	}
	l.outputLogger.Println(string(bizlogData))
	err = l.repo.InsertLogInDB(l.ctx, &bizlog)
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

// LogUpdate records changes to an entity.
//
// Arguments:
//   - entity represents the type of business entity being modified. (user/context)
//   - username of who performed the action.
//   - userRole represents the role of the user performing the action. (SA/CA)
//   - context provides additional business context (optional).
//   - entityID identifies uniquely the affected entity.
//   - newValue contains the new state of the entity.
//   - oldValue contains the previous state of the entity.
//   - description provides additional information or explanation of the event (optional).
func (l *Logger) LogUpdate(entity Entity, username string, userRole UserRole, context *string, entityID string, oldValue any, newValue any, description *string) {
	bizlog := BizLog{
		Timestamp:   time.Now().Format("2006-01-02T15:04:05.000Z"),
		EventType:   EventTypeUpdate,
		Entity:      entity,
		Username:    username,
		UserRole:    userRole,
		Context:     context,
		EntityID:    entityID,
		OldValue:    oldValue,
		NewValue:    newValue,
		Description: description,
	}
	bizlogData, err := json.Marshal(bizlog)
	if err != nil {
		l.techLogger.Error("LogUpdate json.Marshal", slog.Any("error", err))
		return
	}
	l.outputLogger.Println(string(bizlogData))
	err = l.repo.InsertLogInDB(l.ctx, &bizlog)
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
