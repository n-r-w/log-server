// Package usecase Модели данных, относящиеся к журналу. Сейчас все в одном файле.
// При большом количестве моделей и операций имеет смысл разбить на несколько файлов или каталогов
package usecase

import (
	"time"

	"github.com/n-r-w/log-server/internal/domain/model"
	"github.com/n-r-w/log-server/internal/repository"
	"github.com/pkg/errors"
)

type logCase struct {
	RepoLog repository.LogInterface
}

func NewLogCase(r repository.LogInterface) LogInterface {
	return &logCase{
		RepoLog: r,
	}
}

func (l *logCase) Insert(logs *[]model.LogRecord) error {
	return errors.Wrap(l.RepoLog.Insert(logs), "insert error")
}

func (l *logCase) Find(dateFrom time.Time, dateTo time.Time) (*[]model.LogRecord, error) {
	r, e := l.RepoLog.Find(dateFrom, dateTo)
	return r, errors.Wrap(e, "find error")
}
