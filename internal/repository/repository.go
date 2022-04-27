// Package repository
// Содержит функционал для чтения/записи данных из внешних источников
package repository

import (
	"time"

	"github.com/n-r-w/log-server/internal/domain/model"
)

type UserInterface interface {
	// Insert добавить нового пользователя. ID прописывается в модель
	Insert(user *model.User) error
	Remove(userID int64) error
	Update(user *model.User) error
	ChangePassword(userID int64, password string) error

	FindByID(userID int64) (*model.User, error)
	FindByLogin(login string) (*model.User, error)
	GetUsers() ([]model.User, error)
}

type LogInterface interface {
	Insert(records *[]model.LogRecord) error

	Find(dateFrom time.Time, dateTo time.Time) (*[]model.LogRecord, error)
}
