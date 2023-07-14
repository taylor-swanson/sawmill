package v1

import (
	"archive/zip"

	"github.com/taylor-swanson/sawmill/internal/component/logs"
	"github.com/taylor-swanson/sawmill/internal/logger"
)

func FindLogs(bundle *viewer) []logs.Entry {
	var entries []logs.Entry

	err := bundle.Walk("logs/", func(file *zip.File) error {
		if file.FileInfo().IsDir() {
			return nil
		}

		entries = append(entries, logs.Entry{
			Filename:  file.Name,
			Type:      logs.GetType(file.Name),
			Component: logs.GetComponent(file.Name),
		})

		return nil
	})
	if err != nil {
		logger.Error().Err(err).Msg("Error getting configs")
	}

	return entries
}
