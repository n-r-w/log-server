// Package psql Содержит реализацию интерфейса репозитория логов для postgresql
package psql

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/n-r-w/log-server/internal/app/config"
	"github.com/n-r-w/log-server/internal/domain/model"
	"github.com/n-r-w/log-server/internal/repository"
	"github.com/pkg/errors"
)

// Релизация интерфейса LogInterface для psql
type logImpl struct {
	dbImpl *sqlDbImpl
	db     *pgxpool.Pool
}

// NewLog Возвращаем интерфейс работы с логами
func NewLog(db repository.DBOInterface) repository.LogInterface { //nolint:ireturn
	dbImpl, ok := db.(*sqlDbImpl)
	if !ok {
		log.Panicln("internal error")
	}

	return &logImpl{
		dbImpl: dbImpl,
		db:     dbImpl.db,
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

func (p *logImpl) Find(dateFrom time.Time, dateTo time.Time, limit int) (records *[]model.LogRecord, limited bool, err error) {
	rows, err := p.db.Query(context.Background(),
		`SELECT id, record_timestamp, real_timestamp, level,  message1, COALESCE(message2, ''), COALESCE(message3, '') 
		FROM log
		WHERE ($1 OR record_timestamp >= $2) AND ($3 OR record_timestamp <= $4)
		ORDER BY record_timestamp DESC
		LIMIT $5`,
		dateFrom.IsZero(), dateFrom, dateTo.IsZero(), dateTo, limit+1)
	if err != nil {
		return nil, false, errors.Wrap(err, "query error")
	}
	defer rows.Close() // освобождаем контекст sql запроса при выходе

	var recs []model.LogRecord

	var rowCount uint64
	limited = false

	for rows.Next() {
		var record model.LogRecord

		if err := rows.Scan(&record.ID, &record.LogTime, &record.RealTime,
			&record.Level, &record.Message1, &record.Message2, &record.Message3); err != nil {
			return nil, false, errors.Wrap(err, "rows scan error")
		}

		rowCount++
		if rowCount > uint64(limit) {
			limited = true
			break
		}

		if rowCount > uint64(config.AppConfig.MaxLogRecordsResult) {
			err := fmt.Errorf("too many records, max %d", config.AppConfig.MaxLogRecordsResult)
			return nil, false, err
		}

		recs = append(recs, record)
	}

	// при rows.Scan может быть ошибка и тогда defer rows.Close() не вызовется
	// поэтому надежнее сделать как defer rows.Close(), так и прямое закрытие здесь
	rows.Close()

	return &recs, limited, nil
}
