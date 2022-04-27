package httprouter

import (
	"net/http"
)

// Реализует интерфейс http.ResponseWriter
// Подменяет собой стандартный http.ResponseWriter и позволяет дополнительно сохранить в нем ошибку
type responseWriter struct {
	http.ResponseWriter
	code int
	err  error
}

func (w *responseWriter) WriteHeader(statusCode int) {
	w.code = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}
