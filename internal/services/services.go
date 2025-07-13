package services

import (
	"context"
	"path/filepath"
	"time"

	"github.com/rogerwesterbo/svennescamping-backend/internal/services/prices"
	"github.com/rogerwesterbo/svennescamping-backend/pkg/consts"
	"github.com/rogerwesterbo/svennescamping-backend/pkg/interfaces"
	"github.com/rogerwesterbo/svennescamping-backend/pkg/logger"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var (
	PriceService             *prices.PriceService
	GlobalTransactionService *TransactionService
	GlobalBackgroundFetcher  *BackgroundFetcher
)

func InitializeServices() {
	// initialize price service
	// For Kubernetes deployment, you might use an environment variable
	csvPath := viper.GetString(consts.PRICES_CSV_PATH)

	// Make sure the path is absolute for consistency
	absPath, err := filepath.Abs(csvPath)
	if err != nil {
		logger.Fatal("Failed to get absolute path: %v", zap.Error(err))
	}

	// Initialize the price service
	PriceService, err = prices.NewPriceService(absPath)
	if err != nil {
		logger.Fatal("Failed to initialize price service: %v", zap.Error(err))
	}
}

// InitializeTransactionServices initializes transaction-related services
func InitializeTransactionServices(
	cache interfaces.Cache,
	transactionRepo interfaces.TransactionRepository,
	stripeClient interfaces.Transactions,
	vippsClient interfaces.Transactions,
	zettleClient interfaces.Transactions,
) {
	// Initialize transaction service
	GlobalTransactionService = NewTransactionService(transactionRepo)

	// Initialize background fetcher with 5-minute interval
	GlobalBackgroundFetcher = NewBackgroundFetcher(
		cache,
		stripeClient,
		vippsClient,
		zettleClient,
		5*time.Minute,
	)

	logger.Info("Transaction services initialized successfully")
}

// StartBackgroundFetching starts the background data fetching from all providers
func StartBackgroundFetching(ctx context.Context) {
	if GlobalBackgroundFetcher != nil {
		logger.Info("Starting background transaction fetching")
		GlobalBackgroundFetcher.Start(ctx)
	} else {
		logger.Warn("Background fetcher not initialized, cannot start fetching")
	}
}

// StopBackgroundFetching stops the background data fetching
func StopBackgroundFetching() {
	if GlobalBackgroundFetcher != nil {
		logger.Info("Stopping background transaction fetching")
		GlobalBackgroundFetcher.Stop()
	}
}

// IsBackgroundFetchingRunning returns true if background fetching is active
func IsBackgroundFetchingRunning() bool {
	if GlobalBackgroundFetcher != nil {
		return GlobalBackgroundFetcher.IsRunning()
	}
	return false
}
