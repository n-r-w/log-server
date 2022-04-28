// Package psql Хранит экземпляр объекта по непосредственной работе с БД postgress
package psql

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/n-r-w/log-server/internal/app/config"
	"github.com/n-r-w/log-server/internal/repository"
	"github.com/pkg/errors"
)

var sqlDB *sqlDbImpl

// Реализация SqlDbInterface для psql
type sqlDbImpl struct {
	db *pgxpool.Pool
}

func CreatePsqlDBO() (repository.DBOInterface, error) {
	sqlDB = new(sqlDbImpl)

	if err := sqlDB.dbConnect(); err != nil {
		return nil, err
	}

	return sqlDB, nil
}

// Close Завершение работы с хранилищем
//goland:noinspection GoUnnecessarilyExportedIdentifiers
func (s *sqlDbImpl) Close() {
	s.db.Close()
}

// Подключение к БД
func (s *sqlDbImpl) dbConnect() error {
	url := config.AppConfig.DatabaseURL
	url = fmt.Sprintf("%s %s=%d %s=%ds", url,
		"pool_max_conns", config.AppConfig.MaxDbSessions,
		"pool_max_conn_idle_time", config.AppConfig.MaxDbSessionIdleTimeSec)

	dbPool, err := pgxpool.Connect(context.Background(), url)
	if err != nil {
		return errors.Wrap(err, "connect error")
	}

	s.db = dbPool

	// пробуем открыть БД
	if err := dbPool.Ping(context.Background()); err != nil {
		return errors.Wrap(err, "ping error")
	}

	return nil
}
