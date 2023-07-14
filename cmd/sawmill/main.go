package main

import (
	"os"

	"github.com/taylor-swanson/sawmill/cmd/sawmill/cli"
	"github.com/taylor-swanson/sawmill/internal/logger"

	_ "github.com/taylor-swanson/sawmill/internal/bundle/v1"
	_ "github.com/taylor-swanson/sawmill/internal/bundle/v2"
	_ "github.com/taylor-swanson/sawmill/internal/component/logs/ndjson"
	_ "github.com/taylor-swanson/sawmill/internal/component/logs/text"
)

func main() {
	defer func() {
		_ = logger.Close()
	}()

	if err := cli.NewRootCmd().Execute(); err != nil {
		logger.Error().Err(err).Msg("Error running command")
		_ = logger.Close()
		os.Exit(1)
	}
}
