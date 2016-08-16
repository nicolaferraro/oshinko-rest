package logging

import (
	"fmt"
	"net/http"

	"github.com/redhatanalytics/oshinko-rest/helpers/uuid"
)

// LogResponseWriter is a wrapper struct which allows us to retain the
// status code generated by ServeHTTP calls.
type logResponseWriter struct {
	writer http.ResponseWriter
	status int
}

func (w *logResponseWriter) Header() http.Header {
	return w.writer.Header()
}

func (w *logResponseWriter) Write(b []byte) (int, error) {
	return w.writer.Write(b)
}

func (w *logResponseWriter) WriteHeader(s int) {
	w.status = s
	w.writer.WriteHeader(s)
}

// AddLoggingHandler will decorate the passed handler with a wrapper which
// will emit log messages whenever a request is received, in addition to
// calling the original handler.
func AddLoggingHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		l := GetLogger()
		reqId, _ := uuid.Uuid()
		reqId = fmt.Sprintf("[req-id %s]", reqId)
		l.Println(reqId, r.Method, r.URL)
		lr := &logResponseWriter{w, 0}
		next.ServeHTTP(lr, r)
		l.Println(reqId, lr.status)
	})
}
