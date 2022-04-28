package main

import (
	"flag"
	"log"

	"github.com/gorilla/sessions"
	"github.com/n-r-w/log-server/internal/app/config"
	"github.com/n-r-w/log-server/internal/domain"
	"github.com/n-r-w/log-server/internal/domain/usecase"
	"github.com/n-r-w/log-server/internal/presentation/httprouter"
	"github.com/n-r-w/log-server/internal/repository"
	"github.com/n-r-w/log-server/internal/repository/psql"
	"github.com/n-r-w/log-server/internal/repository/testrepo"
)

func main() {
	var configPath string
	// описание флагов командной строки
	flag.StringVar(&configPath, "config-path", "config/server.toml", "path to config file")

	// обработка командной строки
	flag.Parse()

	// читаем конфиг
	if err := config.Load(configPath); err != nil {
		log.Fatal(err)
	}

	// создаем экземпляры объектов, реализующих различные интерфейсы

	var userRepo repository.UserInterface

	var logRepo repository.LogInterface

	if false {
		// реальная БД
		dbo, err := psql.CreatePsqlDBO()
		if err != nil {
			log.Fatal(err)
		}

		userRepo = psql.NewUser(dbo)
		logRepo = psql.NewLog(dbo)
	} else {
		// фейковая БД
		dbo, err := testrepo.CreateTestlDBO()
		if err != nil {
			log.Fatal(err)
		}

		userRepo = testrepo.NewUser(dbo)
		logRepo = testrepo.NewLog(dbo)
	}

	// создаем сценарии
	userUsecase := usecase.NewUserCase(userRepo)
	logCase := usecase.NewLogCase(logRepo)

	// инициализируем домен
	dom := domain.NewDomain(logCase, userUsecase)

	// создаем роутер
	sessionStore := sessions.NewCookieStore([]byte(config.AppConfig.SessionEncriptionKey))
	router := httprouter.NewRouter(dom, sessionStore)

	if err := router.Start(); err != nil {
		log.Fatal(err)
	}
}
