package httprouter

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/n-r-w/log-server/internal/app/config"
	"github.com/n-r-w/log-server/internal/app/logger"
	"github.com/n-r-w/log-server/internal/domain"
	"github.com/n-r-w/log-server/internal/domain/model"
	"github.com/pkg/errors"
	"golang.org/x/exp/slices"
)

// Тип для описания ключевых значений параметров, добавляемых в контекст запроса
// в процессе его обработки через middleware
type contextKey int8

const (
	// AppName Имя приложения
	AppName = "LogServer"
	// SessionName Ключ для хранения информации о сессии со стороны пользователя
	SessionName = "logserver"
	// UserIDKeyName Ключ для хранения id пользователя в сессии (в куках)
	UserIDKeyName = "user_id"

	// BinaryFormatHeaderName Имя хедера REST запроса, в котором клиент указывает в каком виде он желает получить ответ
	BinaryFormatHeaderName = "binary-format"
	// BinaryFormatHeaderProtobuf Требуется ответ в формате protobuf
	BinaryFormatHeaderProtobuf = "protobuf"
)

const (
	// Ключ для хранения модели пользователя в контексте запроса после успешной аунтетификации
	ctxKeyUser contextKey = iota
	// Ключ для хранения в контексте запроса уникального номера сессии
	ctxKeyRequestID contextKey = iota
)

// Формат бинарного ответа
type binaryFormat int

const (
	// Формат бинарного ответа (flatbuffers)
	binaryFormatLogs binaryFormat = 1
)

// HTTPRouter Объект роутер
type HTTPRouter struct {
	router       *mux.Router    // Управление маршрутами
	sessionStore sessions.Store // Управление сессиями пользователей
	domain       *domain.Domain // Унтерфейсы доменной области (сценарии)
}

// NewRouter Создание роутера
func NewRouter(domain *domain.Domain, sessionStore sessions.Store) *HTTPRouter {
	r := HTTPRouter{
		router:       mux.NewRouter(),
		sessionStore: sessionStore,
		domain:       domain,
	}
	// инициализация маршрутов для rest api
	r.initRestRoutes()
	// инициализация маршрутов для web запросов
	r.initWebRoutes()

	return &r
}

// Start Запуск на выполнение
func (router *HTTPRouter) Start() error {
	l, err := net.Listen("tcp", config.AppConfig.BindAddr)
	if err != nil {
		return errors.Wrap(err, "listen error")
	}

	// headersOk := handlers.AllowedHeaders([]string{"X-Requested-With"})
	// originsOk := handlers.AllowedOrigins([]string{"*"})
	// methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"})

	// таймауты
	srv := &http.Server{
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		// Handler:      handlers.CORS(originsOk, methodsOk)(router.router),
		Handler: router.router,
	}

	logger.Logger().Infof("%s listening on %s", AppName, config.AppConfig.BindAddr)

	// Начинаем слушать порт в отдельном потоке
	go func() {
		if err := srv.Serve(l); err != nil {
			log.Println(err)
		}
	}()

	// Обрабатываем сигнал SIGINT (Ctrl+C)
	// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) перехватываться не будут
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	// Ждем получения сигнала
	<-c

	// Время ожидания закрытия соединений
	wait := time.Second * 15

	// Создаем контекст для закрытия
	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()

	// Если нет соединений, то сервер закроется сразу, иначе будет ждать закрытия или истечения времени
	_ = srv.Shutdown(ctx)

	// Если нужно ждать завершения других сервисов, то надо запустить srv.Shutdown в горутине
	// и остановиться на <-ctx.Done()

	log.Println("shutting down")

	return nil
}

// Ответ с ошибкой
func (router *HTTPRouter) respondError(w http.ResponseWriter, r *http.Request, code int, err error) {
	router.respond(w, r, code, map[string]string{"error": err.Error()})
}

// Ответ на запрос без сжатия
func (router *HTTPRouter) respond(w http.ResponseWriter, _ *http.Request, code int, data interface{}) {
	if code > 0 {
		w.WriteHeader(code)
	}
	if data != nil {
		switch d := data.(type) {
		case string:
			_, _ = w.Write([]byte(d))
		default:
			if err := json.NewEncoder(w).Encode(data); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(fmt.Sprintf(`{"error": "%v"}`, err)))

				return
			}

			w.Header().Set("Content-Type", "application/json")
		}
	} else {
		_, _ = w.Write([]byte("{}"))
	}
}

// Ответ на запрос со сжатием если его поддерживает клиент
func (router *HTTPRouter) respondCompressed(w http.ResponseWriter, r *http.Request, code int, data interface{}) {
	if data == nil {
		router.respond(w, r, code, data)

		return
	}

	// проверяем хочет ли клиент сжатие
	accepted := strings.Split(r.Header.Get("Accept-Encoding"), ",")
	gzipCompression := slices.Contains(accepted, "gzip")
	deflateCompression := !gzipCompression && slices.Contains(accepted, "deflate")

	if !gzipCompression && !deflateCompression {
		router.respond(w, r, code, data)

		return
	}

	// заполняем буфер для сжатия
	var sourceData []byte
	switch d := data.(type) {
	case string:
		sourceData = []byte(d)
	default:
		var errJSON error
		sourceData, errJSON = json.Marshal(data)

		if errJSON != nil {
			router.respondError(w, r, http.StatusInternalServerError, errJSON)
		}

		w.Header().Set("Content-Type", "application/json")
	}

	if deflateCompression {
		w.Header().Set("Content-Encoding", "deflate")
	} else {
		w.Header().Set("Content-Encoding", "gzip")
	}

	compressedData, err := compressData(deflateCompression, sourceData)

	if err != nil {
		router.respondError(w, r, http.StatusInternalServerError, err)

		return
	}

	w.WriteHeader(code)
	_, _ = w.Write(compressedData)
}

// Ответ на запрос в формате flatbuffers
func (router *HTTPRouter) respondBinary(w http.ResponseWriter, r *http.Request, code int, data interface{},
	format binaryFormat, compress bool,
) {
	if data == nil {
		router.respond(w, r, code, data)

		return
	}

	switch format {
	case binaryFormatLogs:
		w.Header().Add(BinaryFormatHeaderName, BinaryFormatHeaderProtobuf)

	default:
		log.Fatalln("bad binary format")
	}

	if compress {
		dataBytes, ok := data.([]byte)
		if !ok {
			log.Fatalln("internal error")

			return
		}

		compressedBytes, err := compressData(false, dataBytes)

		if err != nil {
			router.respondError(w, r, http.StatusInternalServerError, err)

			return
		}

		w.WriteHeader(code)
		_, _ = w.Write(compressedBytes)

		return
	}

	w.Header().Add("Content-Type", "application/octet-stream")

	switch d := data.(type) {
	case []byte:
		_, _ = w.Write(d)
	default:
		log.Fatalln("bad binary data")
	}
}

// Текущий пользователь. Он помещается в контекст в методе setRequestID
func currentUser(r *http.Request) *model.User {
	user, ok := r.Context().Value(ctxKeyUser).(*model.User)
	if ok {
		return user
	}

	log.Fatalln("internal error")

	return nil
}

// Сжатие массива данных
func compressData(deflateCompression bool, data []byte) (resData []byte, err error) {
	if data == nil {
		return []byte{}, nil
	}

	// алгоритм сжатия
	var compressor io.WriteCloser
	// целевой буфер
	var compressedBuf bytes.Buffer

	// сжимаем по нужному алгоритму
	if deflateCompression {
		if compressor, err = flate.NewWriter(&compressedBuf, flate.BestSpeed); err != nil {
			return nil, errors.Wrap(err, "deflate error")
		}
	} else {
		if compressor, err = gzip.NewWriterLevel(&compressedBuf, gzip.BestSpeed); err != nil {
			return nil, errors.Wrap(err, "gzip error")
		}
	}

	if _, err := compressor.Write(data); err != nil {
		return nil, errors.Wrap(err, "compress error")
	}

	if err := compressor.Close(); err != nil {
		return nil, errors.Wrap(err, "compress error")
	}

	return compressedBuf.Bytes(), nil
}
