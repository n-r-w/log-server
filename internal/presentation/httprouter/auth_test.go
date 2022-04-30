package httprouter_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/n-r-w/log-server/internal/app/config"
	"github.com/n-r-w/log-server/internal/domain"
	"github.com/n-r-w/log-server/internal/domain/model"
	"github.com/n-r-w/log-server/internal/domain/usecase"
	"github.com/n-r-w/log-server/internal/presentation/httprouter"
	"github.com/n-r-w/log-server/internal/repository"
	"github.com/n-r-w/log-server/internal/repository/testrepo"
	"github.com/stretchr/testify/assert"
)

func initAuthTestCase(t *testing.T) (*httprouter.HTTPRouter, repository.UserInterface, repository.LogInterface) {
	t.Helper()
	assert.NoError(t, config.Load(""))

	// фейковая БД
	dbo, err := testrepo.CreateTestlDBO()
	assert.NoError(t, err)

	// репозитории
	userRepo := testrepo.NewUser(dbo)
	logRepo := testrepo.NewLog(dbo)

	// сценарии
	userUsecase := usecase.NewUserCase(userRepo)
	logCase := usecase.NewLogCase(logRepo)

	// инициализируем домен
	dom := domain.NewDomain(logCase, userUsecase)

	// создаем роутер
	return httprouter.NewRouter(dom, sessions.NewCookieStore([]byte(config.AppConfig.SessionEncriptionKey))), userRepo, logRepo
}

func TestHTTPRouter_AuthenticateUser(t *testing.T) {
	// инициализируем все что надо
	router, userRepo, _ := initAuthTestCase(t)

	// тестовый юзер
	u := model.TestUser(t)
	// заносим его в фейковую БД
	assert.NoError(t, userRepo.Insert(u))

	testCases := []struct {
		name         string
		cookieValue  map[interface{}]interface{}
		expectedCode int
	}{
		{
			name: "authenticated",
			cookieValue: map[interface{}]interface{}{
				httprouter.UserIDKeyName: u.ID,
			},
			expectedCode: http.StatusOK,
		},
		{
			name:         "not authenticated",
			cookieValue:  nil,
			expectedCode: http.StatusUnauthorized,
		},
	}

	// создаем новый куки
	sc := securecookie.New([]byte(config.AppConfig.SessionEncriptionKey), nil)
	// создаем функцию аутентификации
	mw := router.AuthenticateUser(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for _, tc := range testCases { //nolint:paralleltest
		t.Run(tc.name, func(t *testing.T) {
			// создаем тестовую реализацию http.ResponseWriter
			rec := httptest.NewRecorder()
			// создаем новый запрос
			req, _ := http.NewRequest(http.MethodGet, "/api/private/whoami", nil)
			// имитируем наличие в куки номера сессии или его отсутствие (в зависимости от теста)
			cookieStr, _ := sc.Encode(httprouter.SessionName, tc.cookieValue)
			// запихиваем в заголовок эти куки
			req.Header.Set("Cookie", fmt.Sprintf("%s=%s", httprouter.SessionName, cookieStr))
			// вызываем функцию аутентификации через имитацию http запроса
			mw.ServeHTTP(rec, req)
			// проверяем что вышло
			assert.Equal(t, tc.expectedCode, rec.Code)
		})
	}
}
