package bizlogger

import (
	"bizlogger/bizlogger/postgres"
	"context"
	"log"
	"log/slog"
	"os"
	"sync"
)

type Logger struct {
	ctx          context.Context
	repo         *PostgresRepo
	outputLogger *log.Logger
	file         *os.File
	techLogger   *slog.Logger
	wg           *sync.WaitGroup
	queue        []BizLog
	stopChan     chan struct{}
}

func New(ctx context.Context, pgUrl string) (*Logger, error) {
	techLogger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	outputLogger := log.New(os.Stdout, "", 0)
	file, err := os.Create("bizlog.json")
	if err != nil {
		techLogger.Error("bizlogger New", slog.Any("error", err))
		return nil, err
	}
	pg, err := postgres.NewPostgres(pgUrl)
	if err != nil {
		techLogger.Error("bizlogger New", slog.Any("error", err))
		return nil, err
	}
	repo := newPostgresRepo(pg)
	err = repo.createTables("bizlogger/query.sql")
	if err != nil {
		techLogger.Error("bizlogger New", slog.Any("error", err))
		return nil, err
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
	go l.processFile()
	go l.processQueue()

	return l, nil
}

func (l *Logger) Close() {
	close(l.stopChan)
	l.wg.Wait()
}
