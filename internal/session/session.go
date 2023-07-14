package session

import (
	"github.com/google/uuid"

	"github.com/taylor-swanson/sawmill/internal/bundle"
	"github.com/taylor-swanson/sawmill/internal/component/logs"
)

type Session struct {
	ID               uuid.UUID
	Filename         string
	OriginalFilename string
	Hash             string
	Viewer           bundle.Viewer
	LogContexts      map[string]*logs.Context
}
