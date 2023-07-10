package api

import (
	"context"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/multierr"

	"github.com/taylor-swanson/sawmill/internal/bundle"
	"github.com/taylor-swanson/sawmill/internal/logger"
	"github.com/taylor-swanson/sawmill/internal/ui"
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

	indexTmpl         *template.Template
	bundleDetailsFrag *template.Template

	maxUploadSize int64
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

func (h *Handler) handleGetRoot(w http.ResponseWriter, r *http.Request) {
	if err := h.indexTmpl.ExecuteTemplate(w, "base", nil); err != nil {
		PropsFromContext(r.Context()).AppendError(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusTeapot)
}

func (h *Handler) handlePostUpload(w http.ResponseWriter, r *http.Request) {
	type BundleDetails struct {
		Filename  string
		BuildTime string
		Commit    string
		Snapshot  bool
		Version   string
		ID        string
	}

	if err := r.ParseMultipartForm(h.maxUploadSize); err != nil {
		// TODO: Add nicer error handling.
		PropsFromContext(r.Context()).AppendError(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		// TODO: Add nicer error handling.
		PropsFromContext(r.Context()).AppendError(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// TODO: Save or cache bundle. For now, just quickly read it and return some basic info.
	tmpFile, err := os.CreateTemp("", "sawmill-*.zip")
	if err != nil {
		PropsFromContext(r.Context()).AppendError(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	logger.Debug().Str("path", tmpFile.Name()).Str("bundle_filename", header.Filename).Msg("Writing bundle to temporary file")
	defer func() {
		if err := os.Remove(tmpFile.Name()); err != nil {
			logger.Warn().Err(err).Str("path", tmpFile.Name()).Msg("Failed to remove bundle file")
		}
	}()
	if _, err = io.Copy(tmpFile, file); err != nil {
		PropsFromContext(r.Context()).AppendError(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	_ = tmpFile.Close()

	viewer, err := bundle.NewViewer(tmpFile.Name())
	if err != nil {
		PropsFromContext(r.Context()).AppendError(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	bundleDetails := BundleDetails{
		Filename:  header.Filename,
		BuildTime: viewer.Info().BuildTime.Format(time.RFC3339),
		Commit:    viewer.Info().Commit,
		Snapshot:  viewer.Info().Snapshot,
		Version:   viewer.Info().Version,
		ID:        viewer.Info().ID,
	}

	if err = h.bundleDetailsFrag.Execute(w, &bundleDetails); err != nil {
		// TODO: Add nicer error handling.
		PropsFromContext(r.Context()).AppendError(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h *Handler) loadTemplates() error {
	var err error

	h.indexTmpl, err = template.ParseFS(ui.FS, "templates/layouts/*.gohtml", "templates/index.gohtml")
	if err != nil {
		return fmt.Errorf("unable to parse template: %w", err)
	}
	h.bundleDetailsFrag, err = template.ParseFS(ui.FS, "templates/fragments/bundle_details.gohtml")
	if err != nil {
		return fmt.Errorf("unable to parse template: %w", err)
	}

	return nil
}

func NewHandler() (http.Handler, error) {
	h := &Handler{
		Mux:           chi.NewRouter(),
		maxUploadSize: 100 * 1024 * 1024, // 100 MB
	}
	h.Use(
		h.middlewareRecovery,
		h.middlewareCtxProps,
		h.middlewareLogger,
		middleware.StripSlashes,
	)

	if err := h.loadTemplates(); err != nil {
		return nil, err
	}

	// Routes
	h.Get("/", h.handleGetRoot)
	h.Post("/upload", h.handlePostUpload)

	return h, nil
}
