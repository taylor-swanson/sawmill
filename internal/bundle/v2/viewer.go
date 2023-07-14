package v2

import (
	"archive/zip"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"time"

	"github.com/taylor-swanson/sawmill/internal/bundle"
	"github.com/taylor-swanson/sawmill/internal/component/config"
	"github.com/taylor-swanson/sawmill/internal/component/logs"
)

const (
	Name = "v2"

	versionFile = "version.txt"
)

type viewer struct {
	zr       *zip.ReadCloser
	info     bundle.Info
	filename string

	configs []config.Entry
	logs    []logs.Entry
}

func (b *viewer) Info() bundle.Info {
	return b.info
}

func (b *viewer) Close() error {
	return b.zr.Close()
}

func (b *viewer) String() string {
	if b.info.Snapshot {
		return fmt.Sprintf("%s (version %s SNAPSHOT [commit: %s build date: %s] filetype: %s)", filepath.Base(b.filename), b.info.Version, b.info.Commit, b.info.BuildTime.Format(time.RFC3339), Name)
	}

	return fmt.Sprintf("%s (version: %s [commit: %s build date: %s] filetype: %s)", filepath.Base(b.filename), b.info.Version, b.info.Commit, b.info.BuildTime.Format(time.RFC3339), Name)
}

func (b *viewer) OpenFile(filename string) (fs.File, error) {
	f, err := b.zr.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("unable to open file %q: %w", filename, err)
	}

	return f, nil
}

func (b *viewer) Walk(dirname string, walkFn func(file *zip.File) error) error {
	for _, v := range b.zr.File {
		if !strings.HasPrefix(v.Name, dirname) {
			continue
		}

		if err := walkFn(v); err != nil {
			return err
		}
	}

	return nil
}

func (b *viewer) GetConfigs() []config.Entry {
	return b.configs
}

func (b *viewer) GetLogs() []logs.Entry {
	return b.logs
}

func New(filename string) (bundle.Viewer, error) {
	var err error

	b := viewer{filename: filename}

	if b.zr, err = zip.OpenReader(filename); err != nil {
		return nil, fmt.Errorf("unable to create new bundle: %w", err)
	}
	infoReader, err := b.zr.Open(versionFile)
	if err != nil {
		_ = b.zr.Close()
		return nil, fmt.Errorf("unable to read bundle info: %w", err)
	}
	defer infoReader.Close()

	b.info, err = bundle.ParseInfo(infoReader)
	if err != nil {
		_ = b.zr.Close()
		return nil, fmt.Errorf("unable to parse bundle info: %w", err)
	}

	return &b, nil
}

func init() {
	if err := bundle.Register(Name, bundle.ViewerSpec{
		DetectFn:  Detect,
		FactoryFn: New,
	}); err != nil {
		panic(err)
	}
}
