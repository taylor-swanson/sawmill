package v1

import (
	"archive/zip"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"time"

	"github.com/taylor-swanson/sawmill/internal/bundle"
)

const (
	Name = "v1"
)

var versionFilepaths = []string{
	"meta/elastic-agent-version.yaml",
	"meta/elastic-agent-versionyaml",
}

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
	infoReader, err := openVersionFile(b.zr)
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

func openVersionFile(reader *zip.ReadCloser) (fs.File, error) {
	var file fs.File

	for _, v := range versionFilepaths {
		file, _ = reader.Open(v)
	}

	if file == nil {
		return nil, errors.New("unable to find version file")
	}

	return file, nil
}

func init() {
	if err := bundle.Register(Name, bundle.ViewerSpec{
		DetectFn:  Detect,
		FactoryFn: New,
	}); err != nil {
		panic(err)
	}
}
