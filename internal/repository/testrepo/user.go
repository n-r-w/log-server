package testrepo

import (
	"log"
	"strings"

	"github.com/n-r-w/log-server/internal/app/config"
	"github.com/n-r-w/log-server/internal/domain/model"
	"github.com/n-r-w/log-server/internal/repository"
	werrors "github.com/pkg/errors"
)

// Релизация интерфейса UserInterface для psql
type testUserImpl struct {
	dbImpl *testDbImpl
} //nolint:nolintlint,ireturn

// NewUser Возвращаем интерфейс работы с пользователем
func NewUser(db repository.DBOInterface) repository.UserInterface { //nolint:ireturn
	dbImpl, ok := db.(*testDbImpl)
	if !ok {
		log.Panicln("internal error")
	}

	return &testUserImpl{
		dbImpl: dbImpl,
	}
}

// Insert Добавить нового пользвателя
func (r *testUserImpl) Insert(user *model.User) error {
	if user.ID == config.AppConfig.SuperAdminID || strings.EqualFold(user.Login, config.AppConfig.SuperAdminLogin) {
		return repository.ErrCantChangeAdminUser
	}

	if err := user.Prepare(true); err != nil {
		return werrors.Wrap(err, "user prepare error")
	}

	if err := user.Validate(); err != nil {
		return werrors.Wrap(err, "user validate error")
	}

	r.dbImpl.userMutex.Lock()
	r.dbImpl.userByID[r.dbImpl.userIdMax] = user
	r.dbImpl.userIdMax++
	user.ID = r.dbImpl.userIdMax
	r.dbImpl.userMutex.Unlock()

	return nil
}

// ChangePassword Изменить пароль пользователя
func (r *testUserImpl) ChangePassword(userID uint64, password string) error {
	if userID == config.AppConfig.SuperAdminID {
		return repository.ErrCantChangeAdminPassword
	}

	password = strings.TrimSpace(password)

	user, err := r.FindByID(userID)

	if err != nil {
		return err
	}

	if user == nil {
		return repository.ErrUserNotFound
	}

	user.Password = password
	if err = user.Validate(); err != nil {
		return werrors.Wrap(err, "Validate error")
	}

	return nil
}

// FindByID Поиск пользователя по ID
func (r *testUserImpl) FindByID(userID uint64) (*model.User, error) {
	// не админ ли это?
	if userID == config.AppConfig.SuperAdminID {
		return model.AdminUser(), nil
	}

	return r.dbImpl.userByID[userID], nil
}

// FindByLogin Поиск пользователя по логину
func (r *testUserImpl) FindByLogin(login string) (*model.User, error) {

	// не админ ли это?
	if strings.EqualFold(login, config.AppConfig.SuperAdminLogin) {
		return model.AdminUser(), nil
	}

	for _, u := range r.dbImpl.userByID {
		if u.Login == login {
			return u, nil
		}
	}

	return nil, nil
}

// GetUsers Получить список пользователей
func (r *testUserImpl) GetUsers() (*[]model.User, error) {
	users := make([]model.User, 0, len(r.dbImpl.userByID))
	for _, u := range r.dbImpl.userByID {
		users = append(users, *u)
	}

	return &users, nil
}

func (r *testUserImpl) Remove(id uint64) error {
	if u, _ := r.FindByID(id); u == nil {
		return repository.ErrUserNotFound
	}

	delete(r.dbImpl.userByID, id)

	return nil
}

func (r *testUserImpl) Update(user *model.User) error {
	if u, _ := r.FindByID(user.ID); u == nil {
		return repository.ErrUserNotFound
	}

	r.dbImpl.userByID[user.ID] = user

	return nil
}
