// Package psql Содержит интерфейс для работы с хранилищем пользователей в postgres
package psql

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"strings"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/n-r-w/log-server/internal/app"
	"github.com/n-r-w/log-server/internal/app/config"
	"github.com/n-r-w/log-server/internal/domain/model"
	"github.com/n-r-w/log-server/internal/repository"
	"github.com/n-r-w/log-server/internal/tool"
	"github.com/omeid/pgerror"
	werrors "github.com/pkg/errors"
)

// Релизация интерфейса UserInterface для psql
type userImpl struct {
	// Здесь находится экземпляр, а не указатель, т.к. все репозитории имеют свой автономный интерфейс
	// это означает, что они могут использовать как одну физическую БД, так и разные
	// За инициализацию бд отвечает фабрика репозиториев
	db *pgxpool.Pool
} //nolint:nolintlint,ireturn

// NewUser Возвращаем интерфейс работы с пользователем
func NewUser(db *pgxpool.Pool) repository.UserInterface { //nolint:ireturn
	return &userImpl{
		db: db,
	}
}

// Фейковый пользователь - админ
func (r *userImpl) adminUser() *model.User {
	user := &model.User{
		ID:                config.AppConfig.SuperAdminID,
		Name:              "admin",
		Login:             config.AppConfig.SuperAdminLogin,
		Password:          config.AppConfig.SuperPassword,
		EncryptedPassword: "",
	}

	if err := tool.LogIf(user.Prepare(true), "internal error"); err != nil {
		log.Fatalln("internal error")

		return nil
	}

	return user
}

// Insert Добавить нового пользвателя
func (r *userImpl) Insert(user *model.User) error {
	if user.ID == config.AppConfig.SuperAdminID || strings.EqualFold(user.Login, config.AppConfig.SuperAdminLogin) {
		return errCantChangeAdminUser
	}

	if err := user.Prepare(true); err != nil {
		return werrors.Wrap(err, "user prepare error")
	}

	if err := user.Validate(); err != nil {
		return werrors.Wrap(err, "user validate error")
	}

	err := r.db.QueryRow(context.Background(),
		"INSERT INTO users (login, name, encrypted_password) VALUES ($1, $2, $3) RETURNING id",
		user.Login,
		user.Name,
		user.EncryptedPassword,
	).Scan(&user.ID)
	if err != nil {
		if e := pgerror.UniqueViolation(err); e != nil {
			return errLoginExist
		}

		return werrors.Wrap(err, "QueryRow error")
	}

	return werrors.Wrap(err, "QueryRow error")
}

// ChangePassword Изменить пароль пользователя
func (r *userImpl) ChangePassword(userID int64, password string) error {
	if userID == config.AppConfig.SuperAdminID {
		return errCantChangeAdminPassword
	}

	password = strings.TrimSpace(password)
	enc, err := tool.EncryptPassword(password)

	if err != nil {
		return werrors.Wrap(err, "EncryptPassword error")
	}

	var user *model.User
	user, err = r.FindByID(userID)

	if err != nil {
		return err
	}

	if user == nil {
		return errUserNotFound
	}

	user.Password = password
	if err = user.Validate(); err != nil {
		return werrors.Wrap(err, "Validate error")
	}

	if err = user.Prepare(true); err != nil {
		return werrors.Wrap(err, "Prepare error")
	}

	_, err = r.db.Exec(context.Background(), "UPDATE users SET encrypted_password=$1 WHERE id=$2", enc, userID)
	if err != nil {
		if e := pgerror.UniqueViolation(err); e != nil {
			return errLoginExist
		}

		return werrors.Wrap(err, "Exec error")
	}

	return nil
}

// FindByID Поиск пользователя по ID
func (r *userImpl) FindByID(userID int64) (*model.User, error) {
	// не админ ли это?
	if userID == config.AppConfig.SuperAdminID {
		return r.adminUser(), nil
	}

	u := &model.User{
		ID:                0,
		Login:             "",
		Name:              "",
		Password:          "",
		EncryptedPassword: "",
	}
	if err := r.db.QueryRow(context.Background(),
		"SELECT id, login, name, encrypted_password FROM users WHERE id = $1",
		userID,
	).Scan(
		&u.ID,
		&u.Login,
		&u.Name,
		&u.EncryptedPassword,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil //nolint:nilnil
		}

		return nil, werrors.Wrap(err, "QueryRow error")
	}

	return u, nil
}

// FindByLogin Поиск пользователя по логину
func (r *userImpl) FindByLogin(login string) (*model.User, error) {
	u := &model.User{
		ID:                0,
		Login:             "",
		Name:              "",
		Password:          "",
		EncryptedPassword: "",
	}

	// не админ ли это?
	if strings.EqualFold(login, config.AppConfig.SuperAdminLogin) {
		u = r.adminUser()
	} else {
		if err := r.db.QueryRow(context.Background(),
			"SELECT id, login, name, encrypted_password FROM users WHERE login = $1",
			login,
		).Scan(
			&u.ID,
			&u.Login,
			&u.Name,
			&u.EncryptedPassword,
		); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, nil // nolint:nilnil
			}

			return nil, werrors.Wrap(err, "QueryRow error")
		}
	}

	return u, nil
}

// GetUsers Получить список пользователей
func (r *userImpl) GetUsers() ([]model.User, error) {
	rows, err := r.db.Query(context.Background(),
		`SELECT id, login, name, encrypted_password FROM users`)
	if err != nil {
		return nil, werrors.Wrap(err, "query error")
	}
	defer rows.Close() // освобождаем контекст sql запроса при выходе

	var users []model.User

	for rows.Next() {
		var usr model.User
		err = rows.Scan(&usr.ID, &usr.Login, &usr.Name, &usr.EncryptedPassword)

		if err != nil {
			return nil, werrors.Wrap(err, "rows scan error")
		}

		users = append(users, usr)
	}

	rows.Close()

	return users, nil
}

func (r *userImpl) Remove(_ int64) error {
	return app.ErrNotImplemented
}

func (r *userImpl) Update(_ *model.User) error {
	return app.ErrNotImplemented
}
