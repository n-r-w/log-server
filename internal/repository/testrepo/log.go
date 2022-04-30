package testrepo

import (
	"log"
	"time"

	"github.com/n-r-w/log-server/internal/domain/model"
	"github.com/n-r-w/log-server/internal/repository"
)

// Релизация интерфейса LogInterface для psql
type testLogImpl struct {
	dbImpl *testDbImpl
}

// NewLog Возвращаем интерфейс работы с логами
func NewLog(db repository.DBOInterface) repository.LogInterface { //nolint:ireturn
	dbImpl, ok := db.(*testDbImpl)
	if !ok {
		log.Panicln("internal error")
	}

	return &testLogImpl{
		dbImpl: dbImpl,
	}
}

func (p *testLogImpl) Insert(records *[]model.LogRecord) error {
	p.dbImpl.logMutex.Lock()
	for _, record := range *records {
		// если тут не делать копию, то в мапе всегда окажется последняя запись
		rcopy := record
		p.dbImpl.logByID[p.dbImpl.logIdMax] = &rcopy
		p.dbImpl.logIdMax++
		rcopy.ID = p.dbImpl.logIdMax
	}
	p.dbImpl.logMutex.Unlock()

	return nil
}

func (p *testLogImpl) Find(dateFrom time.Time, dateTo time.Time, limit int) (records *[]model.LogRecord, limited bool, err error) {
	recs := make([]model.LogRecord, 0, 100)

	p.dbImpl.logMutex.RLock()
	for _, r := range p.dbImpl.logByID {
		if !dateFrom.IsZero() && r.LogTime.Before(dateFrom) {
			continue
		}

		if !dateTo.IsZero() && r.LogTime.After(dateTo) {
			continue
		}

		recs = append(recs, *r)
		if len(recs) >= limit {
			break
		}
	}
	p.dbImpl.logMutex.RUnlock()

	return &recs, false, nil
}
