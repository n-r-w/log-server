package logger

import "github.com/sirupsen/logrus"

var loggerInstance *logrus.Logger // журналирование

// Logger Глобальный логер
func Logger() *logrus.Logger {
	if loggerInstance == nil {
		loggerInstance = logrus.New()
	}

	return loggerInstance
}
