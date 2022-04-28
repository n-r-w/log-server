package model

import (
	"testing"
	"time"
)

// TestUser Валидный тестовый пользователь
func TestUser(t *testing.T) *User {
	t.Helper()

	return &User{
		ID:                10,
		Login:             "testrepo@example.com",
		Name:              "Ivan Petrov",
		Password:          "Qw!12345",
		EncryptedPassword: "",
	}
}

// TestLogRecord Валидная запись лога
func TestLogRecord(t *testing.T) *LogRecord {
	t.Helper()

	return &LogRecord{
		ID:       1,
		LogTime:  time.Now(),
		RealTime: time.Now(),
		Level:    1,
		Message1: "123",
		Message2: "456",
		Message3: "789",
	}
}
