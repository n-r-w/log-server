// Package repository
// Содержит функционал для чтения/записи данных из внешних источников
package repository

import (
	"errors"
	"time"

	"github.com/n-r-w/log-server/internal/domain/model"
)

// DBOInterface Интерфейс объекта доступа к данным. Не содержит методов кроме закрытия, т.к.
// внутри него скрывается конкретная реализация доступа, которая не нужна извне
type DBOInterface interface {
	Close()
}

// UserInterface Интерфейс работы с данными пользователей
type UserInterface interface {
	// Insert добавить нового пользователя. ID прописывается в модель
	Insert(user *model.User) error
	Remove(userID uint64) error
	Update(user *model.User) error
	ChangePassword(userID uint64, password string) error

	FindByID(userID uint64) (*model.User, error)
	FindByLogin(login string) (*model.User, error)
	GetUsers() (*[]model.User, error)
}

type LogInterface interface {
	Insert(records *[]model.LogRecord) error

	Find(dateFrom time.Time, dateTo time.Time) (*[]model.LogRecord, error)
}

var (
	ErrLoginExist              = errors.New("login exist")
	ErrUserNotFound            = errors.New("user not found")
	ErrCantChangeAdminPassword = errors.New("can't change admin password")
	ErrCantChangeAdminUser     = errors.New("can't change admin user")
)
