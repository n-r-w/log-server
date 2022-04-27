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
		w.Header().Add(binaryFormatHeaderName, binaryFormatHeaderProtobuf)

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
