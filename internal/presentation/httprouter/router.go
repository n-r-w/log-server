package httprouter

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"

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
	// Имя приложения
	appName = "LogServer"
	// Ключ для хранения информации о сессии со стороны пользователя
	sessionName = "logserver"
	// Ключ для хранения id пользователя в сессии (в куках)
	userIDKeyName = "user_id"

	// Имя хедера REST запроса, в котором клиент указывает в каком виде он желает получить ответ
	binaryFormatHeaderName = "binary-format"
	// Требуется ответ в формате protobuf
	binaryFormatHeaderProtobuf = "protobuf"
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
func NewRouter(domain *domain.Domain) *HTTPRouter {
	r := HTTPRouter{
		router:       mux.NewRouter(),
		sessionStore: sessions.NewCookieStore([]byte(config.AppConfig.SessionEncriptionKey)),
		domain:       domain,
	}
	// инициализация маршрутов
	r.initRoutes()

	return &r
}

// Start Запуск на выполнение
func (router *HTTPRouter) Start() error {
	l, err := net.Listen("tcp", config.AppConfig.BindAddr)
	if err != nil {
		return errors.Wrap(err, "listen error")
	}

	logger.Logger().Infof("%s listening on %s", appName, config.AppConfig.BindAddr)
	// Запуск event loop TODO - таймауты
	return errors.Wrap(http.Serve(l, router.router), "serve error")
}

// Ответ с ошибкой
func (router *HTTPRouter) respondError(w http.ResponseWriter, r *http.Request, code int, err error) {
	if code > 0 {
		w.WriteHeader(code)
	}

	router.respond(w, r, code, map[string]string{"error": err.Error()})
}

// Ответ на запрос без сжатия
func (router *HTTPRouter) respond(w http.ResponseWriter, _ *http.Request, code int, data interface{}) {
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

			if code > 0 {
				w.WriteHeader(code)
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
	deflateCompression := slices.Contains(accepted, "deflate")

	if !gzipCompression && !deflateCompression {
		router.respond(w, r, code, data)

		return
	}

	// заполняем буфер для сжатия
	var sourceBuf bytes.Buffer
	switch d := data.(type) {
	case string:
		sourceBuf.Write([]byte(d))
	default:
		if err := json.NewEncoder(&sourceBuf).Encode(data); err != nil {
			router.respondError(w, r, http.StatusInternalServerError, err)

			return
		}
		w.Header().Set("Content-Type", "application/json")
	}

	// сжимаем по нужному алгоритму
	var compressor io.WriteCloser

	var compressedBuf bytes.Buffer

	var err error

	if gzipCompression {
		compressor, err = gzip.NewWriterLevel(&compressedBuf, gzip.BestSpeed)

		w.Header().Set("Content-Encoding", "gzip")
	} else if deflateCompression {
		compressor, err = flate.NewWriter(&compressedBuf, flate.BestSpeed)
		w.Header().Set("Content-Encoding", "deflate")
	}

	if err != nil {
		log.Fatal("init compression error")
	}

	if _, err := sourceBuf.WriteTo(compressor); err != nil {
		router.respondError(w, r, http.StatusInternalServerError, err)
		return
	}
	if err := compressor.Close(); err != nil {
		router.respondError(w, r, http.StatusInternalServerError, err)
		return
	}

	w.WriteHeader(code)
	// отдаем результат клиенту
	_, _ = compressedBuf.WriteTo(w)
}

// Ответ на запрос в формате flatbuffers
func (router *HTTPRouter) respondBinary(w http.ResponseWriter, r *http.Request, code int, data interface{},
	format binaryFormat, compress bool,
) {
	if data == nil {
		router.respond(w, r, code, data)

		return
	}

	w.WriteHeader(code)

	switch format {
	case binaryFormatLogs:
		w.Header().Add(binaryFormatHeaderName, binaryFormatHeaderProtobuf)

	default:
		log.Fatalln("bad binary format")
	}

	if compress {
		// жмем
		compressor, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
		if err != nil {
			log.Fatal("init compression error")
		}
		// отдаем результат клиенту
		dataBytes, ok := data.([]byte)
		if ok {
			if _, err := compressor.Write(dataBytes); err != nil {
				router.respondError(w, r, http.StatusInternalServerError, err)

				return
			}
			if err := compressor.Close(); err != nil {
				router.respondError(w, r, http.StatusInternalServerError, err)

				return
			}

			w.Header().Set("Content-Encoding", "gzip")
		} else {
			log.Fatalln("internal error")
		}

		return
	}

	w.Header().Add("Content-Type", "application/octet-stream")

	switch d := data.(type) {
	case []byte:
		_, _ = w.Write(d) //nolint:errcheck
	default:
		log.Fatalln("bad flatbuffers data")
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
