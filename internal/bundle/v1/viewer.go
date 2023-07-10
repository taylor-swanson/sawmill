package v1

import (
	"archive/zip"
	"fmt"
	"path/filepath"
	"time"

	"github.com/taylor-swanson/sawmill/internal/bundle"
)

const (
	Name = "v1"

	versionFile = "meta/elastic-agent-version.yaml"
)

type viewer struct {
	zr       *zip.ReadCloser
	info     bundle.Info
	filename string
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
