// Package model Модели данных, относящиеся к журналу.
// Сейчас все в одном файле. При большом количестве моделей и операций
// имеет смысл разбить на несколько файлов или каталогов
package model

import (
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pkg/errors"
)

type LogRecord struct {
	ID       uint64    `json:"id"`
	LogTime  time.Time `json:"logTime"`
	RealTime time.Time `json:"realTime"`
	Level    uint      `json:"level"`
	Message1 string    `json:"message1"`
	Message2 string    `json:"message2"`
	Message3 string    `json:"message3"`
}

func (l *LogRecord) Validate() error {
	return errors.Wrap(validation.ValidateStruct(
		l,
		validation.Field(&l.LogTime, validation.Required),
		validation.Field(&l.Level, validation.Required),
		validation.Field(&l.Message1, validation.Required),
	), "validation error")
}
