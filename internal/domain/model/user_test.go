package model_test

import (
	"testing"

	"github.com/n-r-w/log-server/internal/app/config"
	"github.com/n-r-w/log-server/internal/domain/model"
	"github.com/n-r-w/log-server/internal/tool"
	"github.com/stretchr/testify/assert"
)

func initUserTestCase(t *testing.T) {
	t.Helper()
	assert.NoError(t, config.Load(""))
}

func TestUser_Validate(t *testing.T) {
	testCases := []struct {
		name    string
		user    func() *model.User
		isValid bool
	}{
		{
			name: "valid",
			user: func() *model.User {
				return model.TestUser(t)
			},
			isValid: true,
		},
		{
			name: "empty Login",
			user: func() *model.User {
				u := model.TestUser(t)
				u.Login = ""
				return u
			},
			isValid: false,
		},
		{
			name: "empty Name",
			user: func() *model.User {
				u := model.TestUser(t)
				u.Name = ""
				return u
			},
			isValid: false,
		},
		{
			name: "empty Password",
			user: func() *model.User {
				u := model.TestUser(t)
				u.Password = ""
				return u
			},
			isValid: false,
		},
	}

	initUserTestCase(t)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.isValid {
				assert.NoError(t, tc.user().Validate())
			} else {
				assert.Error(t, tc.user().Validate())
			}
		})
	}
}

func TestUser_Prepare(t *testing.T) {
	initUserTestCase(t)

	u := model.TestUser(t)

	assert.NoError(t, u.Prepare(false))
	assert.NotEmpty(t, u.EncryptedPassword)

	assert.True(t, tool.ComparePassword(u.EncryptedPassword, u.Password))
}

func TestUser_ComparePassword(t *testing.T) {
	initUserTestCase(t)

	u := model.TestUser(t)
	assert.NoError(t, u.Prepare(false))

	assert.True(t, u.ComparePassword(u.Password))
}
