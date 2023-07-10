package cli

import (
	"context"
	"github.com/spf13/cobra"
	"github.com/taylor-swanson/sawmill/internal/api"
	"github.com/taylor-swanson/sawmill/internal/logger"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func newCmdRun() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run [ARGS]",
		Short: "Run the app",
		RunE:  doRun,
	}

	cmd.Flags().StringP("listen", "l", ":8080", "http server listen address")
	cmd.Flags().BoolP("https", "s", false, "use https")
	cmd.Flags().StringP("cert", "c", "cert.pem", "path to server certificate file")
	cmd.Flags().StringP("key", "k", "key.pem", "path to server key file")

	return cmd
}

func doRun(cmd *cobra.Command, _ []string) error {
	addr, _ := cmd.Flags().GetString("listen")
	cert, _ := cmd.Flags().GetString("cert")
	key, _ := cmd.Flags().GetString("key")
	https, _ := cmd.Flags().GetBool("https")

	handler := api.NewHandler()

	srv := &http.Server{
		Addr:    addr,
		Handler: handler,
	}
	done := make(chan struct{})

	go func() {
		notify := make(chan os.Signal)
		signal.Notify(notify, os.Interrupt)
		<-notify

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		_ = srv.Shutdown(ctx)
		close(done)
	}()

	logger.Info().Str("listen", srv.Addr).Msg("Running sawmill...")
	var err error
	if https {
		err = srv.ListenAndServeTLS(cert, key)
	} else {
		err = srv.ListenAndServe()
	}
	if err != nil && err != http.ErrServerClosed {
		logger.Error().Err(err).Msg("Server error")
		close(done)
	}

	<-done
	logger.Info().Msg("Sawmill shut down")

	return nil
}
