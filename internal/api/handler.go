package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/taylor-swanson/sawmill/internal/collections"
	"html/template"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"go.uber.org/multierr"

	"github.com/taylor-swanson/sawmill/internal/bundle"
	"github.com/taylor-swanson/sawmill/internal/component/config"
	"github.com/taylor-swanson/sawmill/internal/component/logs"
	"github.com/taylor-swanson/sawmill/internal/hash"
	"github.com/taylor-swanson/sawmill/internal/logger"
	"github.com/taylor-swanson/sawmill/internal/session"
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

	indexTmpl *template.Template
	fragments *template.Template

	maxUploadSize int64

	sessions   map[string]*session.Session
	sessionsMu sync.RWMutex
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
		// TODO: Add nicer error handling.
		PropsFromContext(r.Context()).AppendError(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusTeapot)
}

func (h *Handler) handlePostUpload(w http.ResponseWriter, r *http.Request) {
	type SessionState struct {
		Hash             string
		Filename         string
		OriginalFilename string
		Info             bundle.Info
		Configs          []config.Entry
		Logs             []logs.Entry
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

	fileHash, err := hash.SHA256FromReader(file)
	if err != nil {
		// TODO: Add nicer error handling.
		PropsFromContext(r.Context()).AppendError(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	h.sessionsMu.RLock()
	if s, ok := h.sessions[fileHash]; ok {
		logger.Debug().Str("hash", s.Hash).Str("filename", s.Filename).Msg("Using existing session")
		h.sessionsMu.RUnlock()

		state := SessionState{
			Hash:             s.Hash,
			Filename:         s.Filename,
			OriginalFilename: header.Filename,
			Info:             s.Viewer.Info(),
			Configs:          s.Viewer.GetConfigs(),
			Logs:             s.Viewer.GetLogs(),
		}

		if err = h.fragments.ExecuteTemplate(w, "bundleDetail", &state); err != nil {
			// TODO: Add nicer error handling.
			PropsFromContext(r.Context()).AppendError(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		return
	}
	h.sessionsMu.RUnlock()

	// Reset reader back to beginning of file.
	_, _ = file.Seek(0, 0)

	tmpFile, err := os.CreateTemp("", "sawmill-*.zip")
	if err != nil {
		PropsFromContext(r.Context()).AppendError(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	logger.Debug().Str("path", tmpFile.Name()).Str("bundle_filename", header.Filename).Msg("Writing bundle to temporary file")
	//defer func() {
	//	if err := os.Remove(tmpFile.Name()); err != nil {
	//		logger.Warn().Err(err).Str("path", tmpFile.Name()).Msg("Failed to remove bundle file")
	//	}
	//}()
	if _, err = io.Copy(tmpFile, file); err != nil {
		PropsFromContext(r.Context()).AppendError(err)
		w.WriteHeader(http.StatusInternalServerError)
		_ = os.Remove(tmpFile.Name())
		return
	}
	_ = tmpFile.Close()

	viewer, err := bundle.NewViewer(tmpFile.Name())
	if err != nil {
		PropsFromContext(r.Context()).AppendError(err)
		w.WriteHeader(http.StatusInternalServerError)
		_ = os.Remove(tmpFile.Name())
		return
	}

	// Make a new session.
	s := session.Session{
		ID:               uuid.New(),
		Filename:         tmpFile.Name(),
		OriginalFilename: header.Filename,
		Hash:             fileHash,
		Viewer:           viewer,
		LogContexts:      map[string]*logs.Context{},
	}
	logger.Debug().Str("hash", s.Hash).Str("filename", s.Filename).Msg("Creating new session")

	h.sessionsMu.Lock()
	h.sessions[s.Hash] = &s
	h.sessionsMu.Unlock()

	state := SessionState{
		Hash:             s.Hash,
		Filename:         s.Filename,
		OriginalFilename: s.OriginalFilename,
		Info:             s.Viewer.Info(),
		Configs:          s.Viewer.GetConfigs(),
		Logs:             s.Viewer.GetLogs(),
	}

	if err = h.fragments.ExecuteTemplate(w, "bundleDetail", &state); err != nil {
		// TODO: Add nicer error handling.
		PropsFromContext(r.Context()).AppendError(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h *Handler) handleGetInspectConfig(w http.ResponseWriter, r *http.Request) {
	type ConfigInfo struct {
		Filename string
		Content  string
	}

	fileHash := chi.URLParam(r, "hash")
	filename := r.FormValue("filename")

	logger.Debug().Str("hash", fileHash).Str("filename", filename).Msg("Requesting a config file")

	h.sessionsMu.RLock()
	s, ok := h.sessions[fileHash]
	h.sessionsMu.RUnlock()
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	file, err := s.Viewer.OpenFile(filename)
	if err != nil {
		// TODO: Add nicer error handling.
		PropsFromContext(r.Context()).AppendError(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	buf := bytes.NewBuffer(nil)
	if _, err = io.Copy(buf, file); err != nil {
		// TODO: Add nicer error handling.
		PropsFromContext(r.Context()).AppendError(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	configInfo := ConfigInfo{
		Filename: filename,
		Content:  buf.String(),
	}

	if err = h.fragments.ExecuteTemplate(w, "configDetail", &configInfo); err != nil {
		// TODO: Add nicer error handling.
		PropsFromContext(r.Context()).AppendError(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h *Handler) handleGetInspectLog(w http.ResponseWriter, r *http.Request) {
	type LogData struct {
		Fields  []string
		Entries []collections.Fields
	}
	type LogInfo struct {
		Filename  string
		LogData   LogData
		Type      logs.Type
		Component logs.Component
	}

	fileHash := chi.URLParam(r, "hash")
	filename := r.FormValue("filename")

	logger.Debug().Str("hash", fileHash).Str("filename", filename).Msg("Requesting a log file")

	h.sessionsMu.RLock()
	s, ok := h.sessions[fileHash]
	h.sessionsMu.RUnlock()
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	logCtx, ok := s.LogContexts[fileHash]
	if !ok {
		file, err := s.Viewer.OpenFile(filename)
		if err != nil {
			// TODO: Add nicer error handling.
			PropsFromContext(r.Context()).AppendError(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		p, err := logs.NewParser("ndjson")
		if err != nil {
			// TODO: Add nicer error handling.
			PropsFromContext(r.Context()).AppendError(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		logCtx, err = p.Parse(file)
		if err != nil {
			// TODO: Add nicer error handling.
			PropsFromContext(r.Context()).AppendError(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		s.LogContexts[filename] = logCtx
	}

	configInfo := LogInfo{
		Filename: filename,
		LogData: LogData{
			Fields:  logCtx.Fields(),
			Entries: logCtx.ViewAll(),
		},
		Type:      logs.GetType(filename),
		Component: logs.GetComponent(filename),
	}

	if err := h.fragments.ExecuteTemplate(w, "logDetail", &configInfo); err != nil {
		// TODO: Add nicer error handling.
		PropsFromContext(r.Context()).AppendError(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h *Handler) loadTemplates() error {
	var err error

	tmplFuncs := template.FuncMap{
		"configTypeToStr":   func(t config.Type) string { return t.String() },
		"logTypeToStr":      func(t logs.Type) string { return t.String() },
		"logComponentToStr": func(c logs.Component) string { return c.String() },
		"marshalJSON": func(v any) template.JS {
			data, mErr := json.Marshal(v)
			if mErr != nil {
				panic(mErr)
			}
			return template.JS(data)
		},
		"makeTableData": func(entries []collections.Fields, fields []string) template.JS {
			tableData := make([]map[string]any, 0, len(entries))

			i := 1
			for _, entry := range entries {
				entryData := make(map[string]any, len(fields))

				entryData["id"] = i
				for _, field := range fields {
					if field == "id" {
						continue
					}
					if v, ok := entry.Get(field); ok {
						entryData[strings.ReplaceAll(field, ".", "_")] = fmt.Sprintf("%v", v)
					} else {
						entryData[strings.ReplaceAll(field, ".", "_")] = ""
					}
				}
				tableData = append(tableData, entryData)
				i++
			}

			data, mErr := json.Marshal(tableData)
			if mErr != nil {
				panic(mErr)
			}
			return template.JS(data)
		},
		"makeTableColumns": func(fields []string) template.JS {
			type tableColumnData struct {
				Title              string         `json:"title"`
				Field              string         `json:"field"`
				HeaderFilter       string         `json:"headerFilter,omitempty"`
				HeaderFilterParams map[string]any `json:"headerFilterParams,omitempty"`
			}

			tableColumns := make([]tableColumnData, 0, len(fields))
			for _, field := range fields {
				if field == "id" {
					continue
				}
				tcd := tableColumnData{
					Title: field,
					Field: strings.ReplaceAll(field, ".", "_"),
				}
				if field == "log.level" {
					tcd.HeaderFilter = "list"
					tcd.HeaderFilterParams = map[string]any{
						"valuesLookup": true,
						"clearable":    true,
					}
				}
				if field == "" {

				}

				tableColumns = append(tableColumns, tcd)
			}

			data, mErr := json.Marshal(tableColumns)
			if mErr != nil {
				panic(mErr)
			}
			return template.JS(data)
		},
	}

	h.indexTmpl, err = template.New("base").Funcs(tmplFuncs).ParseFS(ui.FS, "templates/layouts/*.gohtml", "templates/index.gohtml")
	if err != nil {
		return fmt.Errorf("unable to parse template: %w", err)
	}
	h.fragments, err = template.New("bundleDetails").Funcs(tmplFuncs).ParseFS(ui.FS, "templates/fragments/*.gohtml")
	if err != nil {
		return fmt.Errorf("unable to parse template: %w", err)
	}

	return nil
}

func (h *Handler) Close() {
	h.sessionsMu.Lock()
	defer h.sessionsMu.Unlock()

	for _, v := range h.sessions {
		_ = v.Viewer.Close()
		if err := os.Remove(v.Filename); err != nil {
			logger.Error().Err(err).Str("filename", v.Filename).Msg("Failed to remove file")
		} else {
			logger.Debug().Str("filename", v.Filename).Msg("Removed file")
		}
	}
}

func NewHandler() (*Handler, error) {
	h := &Handler{
		Mux:           chi.NewRouter(),
		maxUploadSize: 100 * 1024 * 1024, // 100 MB
		sessions:      map[string]*session.Session{},
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
	h.Get("/inspect/config/{hash}", h.handleGetInspectConfig)
	h.Get("/inspect/log/{hash}", h.handleGetInspectLog)

	return h, nil
}
