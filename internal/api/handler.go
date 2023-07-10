package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/multierr"

	"github.com/taylor-swanson/sawmill/internal/logger"
)

// contextKey defines keys for context values.
type contextKey int

const (
	// ctxProps defines a context key for the CtxProps struct.
	ctxProps contextKey = iota + 1
)

// CtxProps defines properties for a context, such as tracking errors.
type CtxProps struct {
	errs error
}

// AppendError appends an error to CtxProps.
func (c *CtxProps) AppendError(err error) {
	c.errs = multierr.Append(c.errs, err)
}

// PropsFromContext retrieves the CtxProps from a context.
func PropsFromContext(ctx context.Context) *CtxProps {
	props, _ := ctx.Value(ctxProps).(*CtxProps)

	return props
}

type Handler struct {
	// Embedding chi.Mux.
	*chi.Mux
}

// middlewareCtxProps injects a CtxProps instance into the request's context.
func (h *Handler) middlewareCtxProps(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r = r.WithContext(context.WithValue(r.Context(), ctxProps, &CtxProps{}))

		next.ServeHTTP(w, r)
	})
}

// middlewareRecovery adds panic recovery to the request.
func (h *Handler) middlewareRecovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			start := time.Now()
			if p := recover(); p != nil && p != http.ErrAbortHandler {
				logger.ErrorStacktrace(p)
				logger.Error().
					Str("panic", fmt.Sprintf("%v", p)).
					Str("method", r.Method).
					Str("url", r.URL.String()).
					Str("client", r.RemoteAddr).
					Dur("elapsed", time.Since(start)).
					Msg("Request panic")
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// middlewareLogger adds logging to the request.
func (h *Handler) middlewareLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		next.ServeHTTP(ww, r)

		props := PropsFromContext(r.Context())
		if props.errs != nil {
			logger.Error().
				Err(props.errs).
				Str("method", r.Method).
				Str("url", r.URL.String()).
				Str("client", r.RemoteAddr).
				Int("status", ww.Status()).
				Int("bytes_written", ww.BytesWritten()).
				Dur("elapsed", time.Since(start)).
				Msg("Request")
		} else {
			logger.Debug().
				Str("method", r.Method).
				Str("url", r.URL.String()).
				Str("client", r.RemoteAddr).
				Int("status", ww.Status()).
				Int("bytes_written", ww.BytesWritten()).
				Dur("elapsed", time.Since(start)).
				Msg("Request")
		}
	})
}

func NewHandler() http.Handler {
	h := &Handler{
		Mux: chi.NewRouter(),
	}
	h.Use(
		h.middlewareRecovery,
		h.middlewareCtxProps,
		h.middlewareLogger,
		middleware.StripSlashes,
	)

	return h
}
