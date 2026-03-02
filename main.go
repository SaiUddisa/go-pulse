package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func healthHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var apis []API
	var jsonRequest Request
	json.NewDecoder(r.Body).Decode(&jsonRequest)
	for _, url := range jsonRequest.URLs {
		apis = append(apis, url)
	}
	doctor := CreateDoctor(WithTimeout(10), WithWorkers(5))
	res := doctor.CheckHealth(ctx, apis)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}
func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	mux := http.NewServeMux()
	mux.HandleFunc("/health-checker", healthHandler)

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	// Start server in goroutine
	go func() {
		fmt.Println("Gopulse Server started at 8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Failed to start server: %v\n", err)
		}
	}()

	// Wait for shutdown signal
	<-ctx.Done()
	fmt.Println("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		fmt.Printf("Server shutdown failed: %v\n", err)
	}

	fmt.Println("Server exited properly")
}
