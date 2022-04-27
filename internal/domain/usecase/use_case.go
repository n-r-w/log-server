// Package usecase Набор возможных бизнес-операций, выполняемых сервисом.
// Сделаны в виде абстракции (интерфейсов).
// Реализация интерфейсов (interactors в терминологии CA) находится в каталоге model
// Степень декомпозиции зависит от того, насколько высокоуровневой является логика.
package usecase

import (
	"errors"
	"time"

	"github.com/n-r-w/log-server/internal/domain/model"
)

// Сейчас операции совпадают с аналогичными интерфейсами в репозитории, но это от того, что такие задачи
// В репозитории должны быть только базовые операции работы с БД, а тут они должны комбинироваться для
// решения составных задач

type UserInterface interface {
	// CheckPassword Проверить пароль
	CheckPassword(login string, password string) (ID int64, err error)
	// ChangePassword Сменить пароль
	ChangePassword(currentUser *model.User, login string, password string) (ID int64, err error)

	Insert(user *model.User) error
	Remove(id int64) error
	Update(user *model.User) error

	FindByID(id int64) (*model.User, error)
	FindByLogin(login string) (*model.User, error)
	GetUsers() ([]model.User, error)
}

type LogInterface interface {
	Insert(logs *[]model.LogRecord) error

	Find(dateFrom time.Time, dateTo time.Time) (*[]model.LogRecord, error)
}

var (
	errNotAdmin     = errors.New("not admin user")
	errUserNotFound = errors.New("user not found")
)
