package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
)

// global logging middleware
func loggerMiddleWare(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := uuid.NewString()
		reqLogger := logger.With(
			slog.String("request_id", requestID),
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
		)

		ctx := withLogger(r.Context(), reqLogger)

		reqLogger.Info("request started")

		next.ServeHTTP(w, r.WithContext(ctx))

		reqLogger.Info("request finished")
	})
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var apis []API
	var jsonRequest Request
	json.NewDecoder(r.Body).Decode(&jsonRequest)
	for _, url := range jsonRequest.URLs {
		apis = append(apis, url)
	}
	doctor := CreateDoctor(WithTimeout(10), WithWorkers(500))
	res := doctor.CheckHealth(ctx, apis)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}
func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	initLogger()
	mux := http.NewServeMux()
	mux.HandleFunc("/health-checker", healthHandler)
	handler := loggerMiddleWare(mux)
	server := &http.Server{
		Addr:    ":8080",
		Handler: handler,
	}

	// Start server in goroutine
	go func() {

		logger.Info("Gopulse Server started at 8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("failed to start server",
				slog.String("error", err.Error()),
			)
		}
	}()

	// Wait for shutdown signal
	<-ctx.Done()
	logger.Info("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("Server shutdown failed",
			slog.String("error", err.Error()),
		)

	}
	logger.Info("Server exited properly")
}
