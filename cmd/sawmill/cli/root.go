package cli

import (
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"

	"github.com/taylor-swanson/sawmill/internal/logger"
)

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sawmill",
		Short: "A tool for examining Elastic Agent diagnostic bundles",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			levelValue, _ := cmd.Flags().GetString("log-level")
			pretty, _ := cmd.Flags().GetBool("log-pretty")

			level, err := zerolog.ParseLevel(levelValue)
			if err != nil {
				return err
			}

			logger.SetupLogger(logger.Options{
				Pretty: pretty,
				Level:  level,
			})

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.AddCommand(
		newCmdRun(),
	)

	cmd.PersistentFlags().StringP("log-level", "L", "info", "set log level (trace, debug, info, warn, error)")
	cmd.PersistentFlags().BoolP("log-pretty", "P", false, "set pretty log formatting")

	return cmd
}
