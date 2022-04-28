package testrepo

import (
	"github.com/n-r-w/log-server/internal/domain/model"
	"github.com/n-r-w/log-server/internal/repository"
)

var testDB *testDbImpl

// Реализация SqlDbInterface для psql
type testDbImpl struct {
	userIdMax uint64
	userByID  map[uint64]*model.User

	logIdMax uint64
	logByID  map[uint64]*model.LogRecord
}

func CreateTestlDBO() (repository.DBOInterface, error) { //nolint:ireturn
	testDB = &testDbImpl{
		userIdMax: 1,
		userByID:  make(map[uint64]*model.User),
		logIdMax:  1,
		logByID:   make(map[uint64]*model.LogRecord),
	}

	return testDB, nil
}

// Close Завершение работы с хранилищем
//goland:noinspection GoUnnecessarilyExportedIdentifiers
func (d *testDbImpl) Close() {
}
