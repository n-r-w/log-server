// Package model Модели данных, относящиеся к пользователю. Сейчас все в одном
// файле. При большом количестве моделей и операций имеет смысл разбить на
// несколько файлов или каталогов Сюда не входит операции, связанные с
// аутентификацие пользователя (логин, выход и т.п.)
package model

import (
	"log"
	"regexp"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	_ "github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/n-r-w/log-server/internal/app/config"
	"github.com/n-r-w/log-server/internal/tool"
	"github.com/pkg/errors"
)

// User Модель пользователя
type User struct {
	ID                uint64 `json:"id"`
	Login             string `json:"login"`
	Name              string `json:"name"`
	Password          string `json:"password,omitempty"`
	EncryptedPassword string `json:"-"`
}

// Validate Валидация ...
func (u *User) Validate() error {
	return validation.ValidateStruct(
		u,
		validation.Field(&u.Login, validation.Required),
		validation.Field(&u.Name, validation.Required),
		validation.Field(&u.Password, validation.When(len(u.EncryptedPassword) == 0, validation.Required)),
		validation.Field(&u.Password, validation.When(len(u.EncryptedPassword) == 0,
			validation.Match(regexp.MustCompile(config.AppConfig.PasswordRegex)).Error(config.AppConfig.PasswordRegexError))),
	)
}

// Prepare Подготовка данных после первой инициализации (инициализация хэша пароля)
func (u *User) Prepare(sanitize bool) error {
	u.Login = strings.TrimSpace(u.Login)
	u.Name = strings.TrimSpace(u.Name)
	u.Password = strings.TrimSpace(u.Password)

	if len(u.Password) > 0 {
		enc, err := tool.EncryptPassword(u.Password)
		if err != nil {
			return errors.Wrap(err, "encript error")
		}

		u.EncryptedPassword = enc
	}

	if sanitize {
		u.sanitize()
	}

	return nil
}

// Очистка пароля после генерации хэша
func (u *User) sanitize() {
	u.Password = ""
}

// ComparePassword Подходит ли пароль
func (u *User) ComparePassword(password string) bool {
	return tool.ComparePassword(u.EncryptedPassword, password)
}

// AdminUser - Фейковый пользователь - админ
func AdminUser() *User {
	user := &User{
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
