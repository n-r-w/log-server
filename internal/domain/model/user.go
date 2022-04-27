// Package model Модели данных, относящиеся к пользователю. Сейчас все в одном
// файле. При большом количестве моделей и операций имеет смысл разбить на
// несколько файлов или каталогов Сюда не входит операции, связанные с
// аутентификацие пользователя (логин, выход и т.п.)
package model

import (
	"fmt"
	"regexp"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/n-r-w/log-server/internal/app/config"
	"github.com/n-r-w/log-server/internal/tool"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

// User Модель пользователя
type User struct {
	ID                int64  `json:"id"`
	Login             string `json:"login"`
	Name              string `json:"name"`
	Password          string `json:"password,omitempty"`
	EncryptedPassword string `json:"-"`
}

// Validate Валидация ...
func (u *User) Validate() error {
	return fmt.Errorf("failed validation %w", validation.ValidateStruct(
		u,
		validation.Field(&u.Login, validation.Required),
		validation.Field(&u.Name, validation.Required),
		validation.Field(&u.Password, validation.By(tool.RequiredIf(u.EncryptedPassword == "")),
			validation.Match(regexp.MustCompile(config.AppConfig.PasswordRegex)).Error(config.AppConfig.PasswordRegexError)),
	))
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
	return bcrypt.CompareHashAndPassword([]byte(u.EncryptedPassword), []byte(password)) == nil
}
