package httprouter

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/n-r-w/log-server/internal/app/config"
	"github.com/n-r-w/log-server/internal/app/logger"
	"github.com/n-r-w/log-server/internal/domain/model"
	"github.com/pkg/errors"
)

var (
	errNotAuthenticated = errors.New("not authenticated")
	errNotAdmin         = errors.New("not admin")
)

// Логин (создание сессии)
func (router *HTTPRouter) handleSessionsCreate() http.HandlerFunc {
	type request struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		loginData := &request{
			Login:    "",
			Password: "",
		}
		// парсим входящий json
		if err := json.NewDecoder(r.Body).Decode(loginData); err != nil {
			router.respondError(w, r, http.StatusBadRequest, err)

			return
		}
		// ищем в БД по логину
		ID, err := router.domain.UserUsecase.CheckPassword(loginData.Login, loginData.Password)
		if err != nil {
			router.respondError(w, r, http.StatusForbidden, err)

			return
		}
		// получаем сесиию
		session, err := router.sessionStore.Get(r, SessionName)
		if err != nil {
			router.respondError(w, r, http.StatusInternalServerError, err)

			return
		}
		// записываем информацию о том, что пользователь с таким ID залогинился
		session.Values[UserIDKeyName] = ID
		session.Options = &sessions.Options{
			Path:   "/",
			Domain: "",
			MaxAge: config.AppConfig.SessionAge,
			Secure: false,
			// HttpOnly: true, // прячем содержимое сессии от доступа через JavaSript в браузере
			HttpOnly: false,
			SameSite: 0,
		}

		if err := router.sessionStore.Save(r, w, session); err != nil {
			router.respondError(w, r, http.StatusInternalServerError, err)

			return
		}

		router.respond(w, r, http.StatusOK, nil)
	}
}

// AuthenticateUser - Аутентификация пользователя на основании ранее прошедшего логина (создания сессии)
func (router *HTTPRouter) AuthenticateUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, httpCode, err := router.isAuthenticated(r)
		if err != nil {
			router.respondError(w, r, httpCode, err)
			return
		}

		// добавляем модель пользователя в контекст запроса
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxKeyUser, user)))
	})
}

// Обработчик запроса с информацией о текущей сессии
func (router *HTTPRouter) handleWhoami() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		router.respond(w, r, http.StatusOK,
			// объект "пользователь" кладется в контекст при логине
			currentUser(r))
	}
}

// Обработчик запроса закрытия сессии
func (router *HTTPRouter) closeSession() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// получаем сесиию
		session, err := router.sessionStore.Get(r, SessionName)
		if err != nil {
			router.respondError(w, r, http.StatusInternalServerError, err)

			return
		}
		// удаляем из нее данные о логине
		delete(session.Values, UserIDKeyName)
		// сохраняем
		if err := router.sessionStore.Save(r, w, session); err != nil {
			logger.Logger().Errorln("session save error")
		}

		router.respond(w, r, http.StatusOK, nil)
	}
}

// Добавить пользователя
func (router *HTTPRouter) addUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if currentUser(r).ID != config.AppConfig.SuperAdminID {
			router.respondError(w, r, http.StatusForbidden, errNotAdmin)

			return
		}

		u := &model.User{
			ID:                0,
			Login:             "",
			Name:              "",
			Password:          "",
			EncryptedPassword: "",
		}
		// парсим входящий json
		if err := json.NewDecoder(r.Body).Decode(u); err != nil {
			router.respondError(w, r, http.StatusBadRequest, err)

			return
		}

		if err := router.domain.UserUsecase.Insert(u); err != nil {
			router.respondError(w, r, http.StatusForbidden, err)
		}

		router.respond(w, r, http.StatusCreated, nil)
	}
}

// Список пользователей
func (router *HTTPRouter) getUsers() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if currentUser(r).ID != config.AppConfig.SuperAdminID {
			router.respondError(w, r, http.StatusForbidden, errNotAdmin)

			return
		}

		users, err := router.domain.UserUsecase.GetUsers()
		if err != nil {
			router.respondError(w, r, http.StatusInternalServerError, err)

			return
		}

		router.respond(w, r, http.StatusOK, &users)
	}
}

// Изменить пароль пользователя
func (router *HTTPRouter) changePassword() http.HandlerFunc {
	type request struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		req := &request{
			Login:    "",
			Password: "",
		}
		// парсим входящий json
		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			router.respondError(w, r, http.StatusBadRequest, err)

			return
		}

		currentUser := currentUser(r)
		if currentUser == nil {
			router.respondError(w, r, http.StatusForbidden, errNotAuthenticated)

			return
		}

		_, err := router.domain.UserUsecase.ChangePassword(currentUser, req.Login, req.Password)
		if err != nil {
			router.respondError(w, r, http.StatusForbidden, err)
		}

		router.respond(w, r, http.StatusOK, nil)
	}
}

func (router *HTTPRouter) isAuthenticated(r *http.Request) (user *model.User, httpCode int, err error) {

	// извлекаем из запроса пользователя куки с инфорамацией о сессии
	session, err := router.sessionStore.Get(r, SessionName)
	if err != nil {
		return nil, http.StatusUnauthorized, err
	}

	// ищем в информацию о пользователе в сессиях
	id, ok := session.Values[UserIDKeyName]
	if !ok || session.Options.MaxAge < 0 {
		return nil, http.StatusUnauthorized, errNotAuthenticated
	}

	// берем инфу о пользователе из БД
	user, err = router.domain.UserUsecase.FindByID(id.(uint64))
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	if user == nil {
		return nil, http.StatusNotFound, errNotAuthenticated
	}

	return user, http.StatusOK, nil
}
