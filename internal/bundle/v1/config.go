package v1

import (
	"archive/zip"
	"path/filepath"
	"strings"

	"github.com/taylor-swanson/sawmill/internal/component/config"
	"github.com/taylor-swanson/sawmill/internal/logger"
)

func GetConfigType(filename string) config.Type {
	filename = filepath.Base(filename)

	if strings.HasPrefix(filename, "elastic-agent-local") {
		return config.TypeAgent
	}
	if strings.HasPrefix(filename, "elastic-agent-policy") {
		return config.TypeAgentPolicy
	}
	if strings.HasPrefix(filename, "endpoint security") {
		return config.TypeEndpoint
	}
	if strings.HasPrefix(filename, "filebeat") {
		return config.TypeFilebeat
	}
	if strings.HasPrefix(filename, "fleet_monitoring") {
		return config.TypeFleetMonitoring
	}
	if strings.HasPrefix(filename, "metricbeat") {
		return config.TypeMetricbeat
	}

	return config.TypeGeneric
}

func FindConfigs(viewer *viewer) []config.Entry {
	var entries []config.Entry

	err := viewer.Walk("config/", func(file *zip.File) error {
		if file.FileInfo().IsDir() {
			return nil
		}

		entries = append(entries, config.Entry{
			Filename: file.Name,
			Type:     GetConfigType(file.Name),
		})

		return nil
	})
	if err != nil {
		logger.Error().Err(err).Msg("Error getting configs")
	}

	return entries
}
