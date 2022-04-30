package httprouter

import (
	"github.com/gorilla/handlers"
)

// инициализация маршрутов
func (router *HTTPRouter) initRestRoutes() {
	// установка middleware
	router.router.Use(router.setRequestID) // подмешивание номера сессии
	router.router.Use(router.logRequest)   // журналирование запросов

	// разрешаем запросы к серверу c любых доменов (cross-origin resource sharing)
	router.router.Use(handlers.CORS(handlers.AllowedOrigins([]string{"*"})))

	// создаем подчиненный роутер для запросов аутентификации
	authSubrout := router.router.PathPrefix("/api/auth").Subrouter()
	// логин
	authSubrout.HandleFunc("/login", router.handleSessionsCreate()).Methods("POST")
	// закрытие сессии
	authSubrout.HandleFunc("/close", router.closeSession()).Methods("DELETE")

	// ========== запросы, которые возможны только после логина ============
	// создаем подчиненный роутер
	private := router.router.PathPrefix("/api/private").Subrouter()
	// устанавливаем middleware для проверки валидности сессии
	private.Use(router.AuthenticateUser)

	// запрос с информацией о текущей сессии
	private.HandleFunc("/whoami", router.handleWhoami())

	// добавить пользователя
	private.HandleFunc("/add-user", router.addUser()).Methods("POST")
	// сменить пароль
	private.HandleFunc("/change-password", router.changePassword()).Methods("PUT")
	// получить список пользователей
	private.HandleFunc("/users", router.getUsers()).Methods("GET")

	// добавить запись в лог
	private.HandleFunc("/add-log", router.addLogRecord()).Methods("POST")
	// получить список записей из лога. Ответ в gzip формате
	private.HandleFunc("/records", router.getLogRecords()).Methods("GET")
}
