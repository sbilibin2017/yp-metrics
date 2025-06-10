package runners

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/sbilibin2017/yp-metrics/internal/logger"
)

type Server interface {
	ListenAndServe() error
	Shutdown(ctx context.Context) error
}

func RunServer(ctx context.Context, srv Server) error {
	errChan := make(chan error, 1)

	go func() {
		logger.Log.Info("Starting HTTP server...")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, context.Canceled) && err != http.ErrServerClosed {
			logger.Log.Errorw("Server ListenAndServe failed", "error", err)
			errChan <- err
		} else {
			logger.Log.Info("Server ListenAndServe exited gracefully")
		}
		close(errChan)
	}()

	select {
	case <-ctx.Done():
		logger.Log.Infow("Shutdown signal received, shutting down server...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			logger.Log.Errorw("Server shutdown error", "error", err)
			return err
		}

		logger.Log.Info("Server shutdown completed gracefully")
		return ctx.Err()

	case err := <-errChan:
		if err != nil {
			logger.Log.Errorw("Server exited with error", "error", err)
		}
		return err
	}
}
