package model_test

import (
	"testing"
	"time"

	"github.com/n-r-w/log-server/internal/app/config"
	"github.com/n-r-w/log-server/internal/domain/model"
	"github.com/stretchr/testify/assert"
)

func initLogTestCase(t *testing.T) {
	t.Helper()
	assert.NoError(t, config.Load(""))
}

func TestLogRecord_Validate(t *testing.T) {
	testCases := []struct {
		name      string
		logRecord func() *model.LogRecord
		isValid   bool
	}{
		{
			name: "valid",
			logRecord: func() *model.LogRecord {
				return model.TestLogRecord(t)
			},
			isValid: true,
		},
		{
			name: "empty LogTime",
			logRecord: func() *model.LogRecord {
				lr := model.TestLogRecord(t)
				lr.LogTime = time.Time{}
				return lr
			},
			isValid: false,
		},
		{
			name: "empty Level",
			logRecord: func() *model.LogRecord {
				lr := model.TestLogRecord(t)
				lr.Level = 0
				return lr
			},
			isValid: false,
		},
		{
			name: "empty Message1",
			logRecord: func() *model.LogRecord {
				lr := model.TestLogRecord(t)
				lr.Message1 = ""
				return lr
			},
			isValid: false,
		},
	}

	initLogTestCase(t)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.isValid {
				assert.NoError(t, tc.logRecord().Validate())
			} else {
				assert.Error(t, tc.logRecord().Validate())
			}
		})
	}
}
