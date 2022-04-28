package config

import (
	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"
)

// Конфиг logserver.toml
type config struct {
	SuperAdminID            uint64
	BindAddr                string `toml:"BIND_ADDR"`
	SuperAdminLogin         string `toml:"SUPERADMIN_LOGIN"`
	SuperPassword           string `toml:"SUPERADMIN_PASSWORD"`
	SessionAge              int    `toml:"SESSION_AGE"`
	LogLevel                string `toml:"LOG_LEVEL"`
	DatabaseURL             string `toml:"DATABASE_URL"`
	SessionEncriptionKey    string `toml:"SESSION_ENCRYPTION_KEY"`
	MaxDbSessions           int    `toml:"MAX_DB_SESSIONS"`
	MaxDbSessionIdleTimeSec int    `toml:"MAX_DB_SESSION_IDLE_TIME_SEC"`
	MaxLogRecordsResult     int64  `toml:"MAX_LOG_RECORDS_RESULT"`
	PasswordRegex           string `toml:"PASSWORD_REGEX"`
	PasswordRegexError      string `toml:"PASSWORD_REGEX_ERROR"`
}

// AppConfig Глобальный конфиг
var AppConfig *config

const (
	superAdminID            = 1
	maxDbSessions           = 50
	maxDbSessionIdleTimeSec = 50
	maxLogRecordsResult     = 100000
	defaultSessionAge       = 60 * 60 * 24 // 24 часа
)

// Load Инициализация конфига значениями по умолчанию
func Load(path string) error {
	AppConfig = &config{
		SuperAdminID:            superAdminID,
		BindAddr:                "http://localhost:8080",
		SuperAdminLogin:         "admin",
		SuperPassword:           "admin",
		SessionAge:              defaultSessionAge,
		LogLevel:                "debug",
		DatabaseURL:             "log",
		SessionEncriptionKey:    "e09469b1507d0e7a98831750aff903e0831a428f9addf3cfa348fa64dcf",
		MaxDbSessions:           maxDbSessions,
		MaxDbSessionIdleTimeSec: maxDbSessionIdleTimeSec,
		MaxLogRecordsResult:     maxLogRecordsResult,
		// PasswordRegex:           "^[A-Za-z0-9@$!%*?&]{8,}$",
		PasswordRegex:      ".*",
		PasswordRegexError: "Латинские буквы, цифры и символы @$!%*?& без пробелов, минимум 4 символа",
	}

	if path == "" {
		return nil
	}
	_, err := toml.DecodeFile(path, AppConfig)

	return errors.Wrap(err, "toml error")
}
