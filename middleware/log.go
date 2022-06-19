package middleware

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/google/uuid"
)

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

// Write overrides Write method of gin.ResponseWriter,
func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)

	n, err := w.ResponseWriter.Write(b)
	if err != nil {
		return 0, fmt.Errorf("middleware.bodyLogWriter.Write: %w", err)
	}

	return n, nil
}

// Logger is custom Gin logger middleware that logs API information to stderr.
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		uuidV4 := uuid.Must(uuid.NewRandom()).String()

		c.Header("X-Transaction-ID", uuidV4)

		bodyBytes, _ := ioutil.ReadAll(c.Request.Body)
		r := ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
		c.Request.Body = r
		req := c.Request
		url := req.URL

		w := &bodyLogWriter{body: &bytes.Buffer{}, ResponseWriter: c.Writer}
		c.Writer = w

		c.Set("context", context.WithValue(context.Background(), model.CtxKeyTransactionID, uuidV4))
		c.Next()

		respTime := time.Since(start)

		go func() {
			apiLogger := log.Logger.With().
				Str("transaction_id", uuidV4).
				Logger()

			logEv := apiLogger.Info()

			if w.Status() >= http.StatusInternalServerError {
				logEv = apiLogger.Error()
			}

			if len(c.Errors.Errors()) > 0 {
				logEv.Err(errors.New(c.Errors.String()))
			}

			if w.Header().Get("Content-Type") == binding.MIMEJSON {
				logEv.Bytes("resp_body", w.body.Bytes())
			}

			logEv.Str("path", url.Path).
				Interface("query", url.Query()).
				Str("method", req.Method).
				Bytes("req_body", bodyBytes).
				Interface("req_header", req.Header).
				Str("proto", req.Proto).
				Str("remote_addr", req.RemoteAddr).
				Interface("resp_header", w.Header()).
				Int("resp_code", w.Status()).
				Dur("resp_time", respTime).
				Msgf("[RESTful-API] %s %s %d", req.Method, url.Path, w.Status())
		}()
	}
}
