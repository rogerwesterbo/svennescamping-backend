package httpserver

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/rogerwesterbo/svennescamping-backend/internal/routes"
	"github.com/rogerwesterbo/svennescamping-backend/pkg/consts"
	"github.com/rogerwesterbo/svennescamping-backend/pkg/logger"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func Start() {
	isDevelopment := viper.GetBool(consts.DEVELOPMENT)

	// Initialize the global logger with colored output
	err := logger.InitLogger(isDevelopment)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
	defer logger.Sync()

	router := mux.NewRouter()
	// Setup routes with logger
	routes.SetupRoutes(router, logger.GetLogger())

	host := "localhost"
	if isDevelopment {
		logger.Info("Running in development mode")
	} else {
		logger.Info("Running in production mode")
		host = ""
	}

	port := "8888"
	url := fmt.Sprintf("%s:%s", host, port)
	server := &http.Server{
		Handler:      router,
		Addr:         url,
		ReadTimeout:  20 * time.Second, // Changed from Millisecond to Second
		WriteTimeout: 20 * time.Second, // Changed from Millisecond to Second
	}

	logger.Info("Starting HTTP server",
		zap.String("address", url),
		zap.String("port", port),
		zap.Bool("development", isDevelopment),
	)

	if err := server.ListenAndServe(); err != nil {
		logger.Fatal("HTTP server failed to start", zap.Error(err))
	}
}
