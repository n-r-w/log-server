// Package usecase Модели данных, относящиеся к пользователю. Сейчас все в одном
// файле. При большом количестве моделей и операций имеет смысл разбить на
// несколько файлов или каталогов Сюда не входит операции, связанные с
// аутентификацие пользователя (логин, выход и т.п.)
package usecase

import (
	"strings"

	"github.com/n-r-w/log-server/internal/app/config"
	"github.com/n-r-w/log-server/internal/domain/model"
	"github.com/n-r-w/log-server/internal/repository"
	"github.com/pkg/errors"
)

type userCase struct {
	UserRepo repository.UserInterface
}

func NewUserCase(r repository.UserInterface) UserInterface {
	return &userCase{
		UserRepo: r,
	}
}

// CheckPassword Проверить пароль
func (u *userCase) CheckPassword(login string, password string) (ID uint64, err error) {
	// ищем в БД по логину
	user, err := u.UserRepo.FindByLogin(login)
	if err != nil {
		return 0, err
	}
	// проверяем наличие пользователя в БД и пароль
	if u == nil || !user.ComparePassword(password) {
		return 0, errors.New("incorrect email or password")
	}

	return user.ID, nil
}

// ChangePassword Проверить пароль
func (u *userCase) ChangePassword(currentUser *model.User, login string, password string) (ID uint64, err error) {
	login = strings.TrimSpace(login)
	password = strings.TrimSpace(password)
	changeSelf := currentUser.Login == login

	var id uint64

	if !changeSelf {
		if currentUser.ID != config.AppConfig.SuperAdminID {
			// если не админ, то менять можно только себе
			return 0, errNotAdmin
		}

		user, err := u.FindByLogin(login)
		if err != nil {
			return 0, err
		}

		if user == nil {
			return 0, errUserNotFound
		}

		id = user.ID
	} else {
		id = currentUser.ID
	}

	return id, errors.Wrap(u.UserRepo.ChangePassword(id, password), "change password error")
}

func (u *userCase) Insert(user *model.User) error {
	return u.UserRepo.Insert(user) //nolint:wrapcheck
}

func (u *userCase) Remove(id uint64) error {
	return u.UserRepo.Remove(id) //nolint:wrapcheck
}

func (u *userCase) Update(user *model.User) error {
	return u.UserRepo.Update(user) //nolint:wrapcheck
}

func (u *userCase) FindByID(id uint64) (*model.User, error) {
	return u.UserRepo.FindByID(id) //nolint:wrapcheck
}

func (u *userCase) FindByLogin(login string) (*model.User, error) {
	return u.UserRepo.FindByLogin(login) //nolint:wrapcheck
}

func (u *userCase) GetUsers() (*[]model.User, error) {
	return u.UserRepo.GetUsers() //nolint:wrapcheck
}
