// Package tool Различные фукции общего назначения
package tool

import (
	"fmt"
	"log"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/n-r-w/log-server/internal/app/logger"
	"golang.org/x/crypto/bcrypt"
)

// EncryptPassword Генерация хэша пароля
func EncryptPassword(s string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(s), bcrypt.MinCost)
	if err != nil {
		return "", fmt.Errorf("failed GenerateFromPassword %w ", err)
	}

	return string(b), nil
}

// RequiredIf Валидатор для проверки по условию
func RequiredIf(cond bool) validation.RuleFunc {
	return func(value interface{}) error {
		if cond {
			return fmt.Errorf("failed validation %w", validation.Validate(value, validation.Required))
		}

		return nil
	}
}

//goland:noinspection GoUnusedExportedFunction,GoUnnecessarilyExportedIdentifiers
func PanicIf(err error, msg string) {
	if err != nil {
		log.Fatalf("error "+msg+": %v", err)
	}
}

func LogIf(err error, msg string) error {
	if err != nil {
		logger.Logger().Errorln(fmt.Sprintf("error %s: %v", msg, err))
	}

	return err
}
