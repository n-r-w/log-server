package main

import (
	"flag"
	"log"

	"github.com/n-r-w/log-server/internal/app/config"
	"github.com/n-r-w/log-server/internal/domain"
	"github.com/n-r-w/log-server/internal/domain/usecase"
	"github.com/n-r-w/log-server/internal/presentation/httprouter"
	"github.com/n-r-w/log-server/internal/repository/psql"
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

	// создаем подключение к БД
	sqlDb, err := psql.NewSQLDb()
	if err != nil {
		log.Fatal(err)
	}

	// создаем экземпляры объектов, реализующих различные интерфейсы
	userRepo := psql.NewUser(sqlDb)
	logRepo := psql.NewLog(sqlDb)

	// создаем сценарии
	userUsecase := usecase.NewUserCase(userRepo)
	logCase := usecase.NewLogCase(logRepo)

	// инициализируем домен
	dom := domain.NewDomain(logCase, userUsecase)

	// создаем роутер
	router := httprouter.NewRouter(dom)

	err = router.Start()
	if err != nil {
		log.Fatal(err)
	}
}
