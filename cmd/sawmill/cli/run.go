package cli

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/spf13/cobra"
	
	"github.com/taylor-swanson/sawmill/internal/api"
	"github.com/taylor-swanson/sawmill/internal/logger"
)

func newCmdRun() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run [ARGS]",
		Short: "Run the app",
		RunE:  doRun,
	}

	cmd.Flags().StringP("listen", "l", ":8082", "http server listen address")
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

	handler, err := api.NewHandler()
	if err != nil {
		return err
	}

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

		if err := srv.Shutdown(ctx); err != nil {
			logger.Error().Err(err).Msg("Server shutdown error")
		}
		close(done)
	}()

	scheme := "http"
	if https {
		scheme = "https"
	}
	_, port, _ := net.SplitHostPort(srv.Addr)

	logger.Debug().Str("listen", srv.Addr).Msg("Started listener")
	logger.Info().Msgf("Access the Sawmill UI at %s://localhost:%s", scheme, port)
	if https {
		err = srv.ListenAndServeTLS(cert, key)
	} else {
		err = srv.ListenAndServe()
	}
	if err != nil && err != http.ErrServerClosed {
		logger.Error().Err(err).Msg("Server error")
	}

	<-done

	handler.Close()

	logger.Debug().Msg("Sawmill shut down")

	return nil
}
