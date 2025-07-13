package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/rogerwesterbo/svennescamping-backend/internal/clients"
	"github.com/rogerwesterbo/svennescamping-backend/internal/httpserver"
	"github.com/rogerwesterbo/svennescamping-backend/internal/middlewares"
	"github.com/rogerwesterbo/svennescamping-backend/internal/services"
	"github.com/rogerwesterbo/svennescamping-backend/internal/settings"
	"github.com/rogerwesterbo/svennescamping-backend/pkg/logger"
	"go.uber.org/automaxprocs/maxprocs"
	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	if err := logger.InitLogger(true); err != nil { // Set to true for development
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
	defer logger.Sync()

	// Set GOMAXPROCS to match Linux container CPU quota
	undo, maxprocsErr := maxprocs.Set()
	if maxprocsErr != nil {
		logger.Error("Failed to set GOMAXPROCS", zap.Error(maxprocsErr))
	} else {
		logger.Info("GOMAXPROCS set successfully")
	}
	defer undo()

	cancelChan := make(chan os.Signal, 1)

	stop := make(chan struct{})
	// catch SIGETRM or SIGINTERRUPT.
	signal.Notify(cancelChan, syscall.SIGTERM, syscall.SIGINT)

	settings.Init()

	// Initialize role service after settings are loaded
	middlewares.InitializeRoleService()

	clients.InitializeClients()

	services.InitializeServices()

	logger.Info("Starting Svennes Camping Backend API")

	// Create context for background operations
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start background transaction fetching
	clients.StartBackgroundFetching(ctx)

	// Start the HTTP server
	go func() {
		httpserver.Start()
		sig := <-cancelChan
		_, _ = fmt.Println()
		_, _ = fmt.Println(sig)
		stop <- struct{}{}
	}()

	logger.Info("Lumi 2025 Backend API started successfully",
		zap.String("version", settings.Version),
		zap.String("commit", settings.Commit),
		zap.String("port", "8888"),
		zap.String("host", "localhost"),
	)
	// Wait for shutdown signal
	<-stop

	// Stop background fetching before shutting down
	logger.Info("Stopping background transaction fetching")
	clients.StopBackgroundFetching()

	logger.Info("Shutting down Lumi 2025 Backend API gracefully",
		zap.String("version", settings.Version),
		zap.String("commit", settings.Commit),
		zap.String("port", "8888"),
		zap.String("host", "localhost"),
	)
}
