package httprouter

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/n-r-w/log-server/internal/app/logger"
	"github.com/sirupsen/logrus"
)

// Добавляем к контексту уникальный ID сесии с ключом ctxKeyRequestID
func (router *HTTPRouter) setRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := uuid.New().String()
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxKeyRequestID, id)))
	})
}

// Выводим все запросы в журнал
func (router *HTTPRouter) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// пишем инфу о начале обработки запроса
		lg := logger.Logger().WithFields(logrus.Fields{
			"remote_addr": r.RemoteAddr,
			"request_id":  r.Context().Value(ctxKeyRequestID),
		})
		lg.Infof("started %s %s", r.Method, r.RequestURI)

		start := time.Now()
		rw := &responseWriter{
			ResponseWriter: w,
			code:           http.StatusOK,
			err:            nil,
		}

		// вызываем обработчик нижнего уровня
		next.ServeHTTP(rw, r)

		// выводим в журнал результат
		var level logrus.Level
		switch {
		case rw.code >= http.StatusInternalServerError:
			level = logrus.ErrorLevel
		case rw.code >= http.StatusBadRequest:
			level = logrus.WarnLevel
		default:
			level = logrus.InfoLevel
		}

		var errorText string
		if rw.err != nil {
			errorText = rw.err.Error()
			errorText = strings.ReplaceAll(errorText, `"`, "")
		} else {
			errorText = "-"
		}

		lg.Logf(
			level,
			`completed with %d %s in %v, info: %s`,
			rw.code,
			http.StatusText(rw.code),
			time.Since(start),
			errorText,
		)
	})
}
