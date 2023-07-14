package bundle

import (
	"archive/zip"
	"io/fs"

	"github.com/taylor-swanson/sawmill/internal/component/config"
	"github.com/taylor-swanson/sawmill/internal/component/logs"
)

type Viewer interface {
	Info() Info
	String() string
	Close() error
	Walk(dirname string, walkFn func(file *zip.File) error) error
	OpenFile(filename string) (fs.File, error)

	GetConfigs() []config.Entry
	GetLogs() []logs.Entry
}
