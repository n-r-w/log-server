// Package psql Содержит реализацию интерфейса репозитория логов для postgresql
package psql

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/n-r-w/log-server/internal/app/config"
	"github.com/n-r-w/log-server/internal/domain/model"
	"github.com/n-r-w/log-server/internal/repository"
	"github.com/pkg/errors"
)

// Релизация интерфейса LogInterface для psql
type logImpl struct {
	// Здесь находится экземпляр, а не указатель, т.к. все репозитории имеют свой автономный интерфейс
	// это означает, что они могут использовать как одну физическую БД, так и разные
	// За инициализацию бд отвечает фабрика репозиториев
	db *pgxpool.Pool
}

func NewLog(db *pgxpool.Pool) repository.LogInterface { //nolint:ireturn
	return &logImpl{
		db: db,
	}
}
func (p *logImpl) Insert(records *[]model.LogRecord) error {
	var sqlText string

	for _, lr := range *records {
		if err := lr.Validate(); err != nil {
			return errors.Wrap(err, "validate error")
		}

		t, _ := lr.LogTime.UTC().MarshalText()
		sqlText += fmt.Sprintf(`INSERT INTO log (record_timestamp, level, message1, message2, message3) 
		 					    VALUES ('%s', %d, '%s', '%s', '%s');`,
			t, lr.Level, lr.Message1, lr.Message2, lr.Message3)
	}

	_, err := p.db.Exec(context.Background(), sqlText)

	return errors.Wrap(err, "exec error")
}

func (p *logImpl) Find(dateFrom time.Time, dateTo time.Time) (*[]model.LogRecord, error) {
	rows, err := p.db.Query(context.Background(),
		`SELECT id, record_timestamp, real_timestamp, level, message1, message2, message3 
		FROM log
		WHERE ($1 OR record_timestamp >= $2) AND ($3 OR record_timestamp <= $4)`,
		dateFrom.IsZero(), dateFrom, dateTo.IsZero(), dateTo)
	if err != nil {
		return nil, errors.Wrap(err, "query error")
	}
	defer rows.Close() // освобождаем контекст sql запроса при выходе

	var records []model.LogRecord

	var rowCount int64

	for rows.Next() {
		var record model.LogRecord

		if err := rows.Scan(&record.ID, &record.LogTime, &record.RealTime,
			&record.Level, &record.Message1, &record.Message2, &record.Message3); err != nil {
			return nil, errors.Wrap(err, "rows scan error")
		}

		rowCount++
		if rowCount > config.AppConfig.MaxLogRecordsResult {
			err := fmt.Errorf("too many records, max %d", config.AppConfig.MaxLogRecordsResult)
			return nil, err
		}

		records = append(records, record)
	}

	// при rows.Scan может быть ошибка и тогда defer rows.Close() не вызовется
	// поэтому надежнее сделать как defer rows.Close(), так и прямое закрытие здесь
	rows.Close()

	return &records, nil
}
