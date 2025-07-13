package services

import (
	"context"
	"sync"
	"time"

	"github.com/rogerwesterbo/svennescamping-backend/pkg/interfaces"
	"github.com/rogerwesterbo/svennescamping-backend/pkg/logger"
	"go.uber.org/zap"
)

type BackgroundFetcher struct {
	cache        interfaces.Cache
	stripeClient interfaces.Transactions
	vippsClient  interfaces.Transactions
	zettleClient interfaces.Transactions
	interval     time.Duration
	stopChan     chan struct{}
	wg           sync.WaitGroup
	running      bool
	mu           sync.RWMutex
}

func NewBackgroundFetcher(
	cache interfaces.Cache,
	stripeClient interfaces.Transactions,
	vippsClient interfaces.Transactions,
	zettleClient interfaces.Transactions,
	interval time.Duration,
) *BackgroundFetcher {
	return &BackgroundFetcher{
		cache:        cache,
		stripeClient: stripeClient,
		vippsClient:  vippsClient,
		zettleClient: zettleClient,
		interval:     interval,
		stopChan:     make(chan struct{}),
	}
}

func (bf *BackgroundFetcher) Start(ctx context.Context) {
	bf.mu.Lock()
	defer bf.mu.Unlock()

	if bf.running {
		logger.Warn("Background fetcher is already running")
		return
	}

	bf.running = true
	logger.Info("Starting background transaction fetcher", zap.Duration("interval", bf.interval))

	// Start goroutines for each provider
	if bf.stripeClient != nil {
		bf.wg.Add(1)
		go bf.fetchFromProvider(ctx, "stripe", bf.stripeClient)
	}

	if bf.vippsClient != nil {
		bf.wg.Add(1)
		go bf.fetchFromProvider(ctx, "vipps", bf.vippsClient)
	}

	if bf.zettleClient != nil {
		bf.wg.Add(1)
		go bf.fetchFromProvider(ctx, "zettle", bf.zettleClient)
	}

	// Initial fetch on startup
	go bf.performInitialFetch(ctx)
}

func (bf *BackgroundFetcher) Stop() {
	bf.mu.Lock()
	defer bf.mu.Unlock()

	if !bf.running {
		return
	}

	logger.Info("Stopping background transaction fetcher")
	close(bf.stopChan)
	bf.wg.Wait()
	bf.running = false
	logger.Info("Background transaction fetcher stopped")
}

func (bf *BackgroundFetcher) IsRunning() bool {
	bf.mu.RLock()
	defer bf.mu.RUnlock()
	return bf.running
}

func (bf *BackgroundFetcher) performInitialFetch(ctx context.Context) {
	logger.Info("Performing initial data fetch from all providers")

	var wg sync.WaitGroup

	// Fetch from all providers in parallel for initial load
	if bf.stripeClient != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			bf.fetchTransactions(ctx, "stripe", bf.stripeClient)
		}()
	}

	if bf.vippsClient != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			bf.fetchTransactions(ctx, "vipps", bf.vippsClient)
		}()
	}

	if bf.zettleClient != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			bf.fetchTransactions(ctx, "zettle", bf.zettleClient)
		}()
	}

	wg.Wait()
	logger.Info("Initial data fetch completed")
}

func (bf *BackgroundFetcher) fetchFromProvider(ctx context.Context, providerName string, client interfaces.Transactions) {
	defer bf.wg.Done()

	ticker := time.NewTicker(bf.interval)
	defer ticker.Stop()

	logger.Info("Started background fetcher for provider", zap.String("provider", providerName))

	for {
		select {
		case <-bf.stopChan:
			logger.Info("Stopping background fetcher for provider", zap.String("provider", providerName))
			return
		case <-ticker.C:
			bf.fetchTransactions(ctx, providerName, client)
		case <-ctx.Done():
			logger.Info("Context cancelled, stopping background fetcher for provider", zap.String("provider", providerName))
			return
		}
	}
}

func (bf *BackgroundFetcher) fetchTransactions(ctx context.Context, providerName string, client interfaces.Transactions) {
	startTime := time.Now()
	logger.Debug("Fetching transactions from provider", zap.String("provider", providerName))

	// Create a timeout context for this specific fetch
	fetchCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	transactions, err := client.GetLatestTransactions(fetchCtx, 100) // Fetch up to 100 transactions
	if err != nil {
		logger.Error("Failed to fetch transactions from provider",
			zap.String("provider", providerName),
			zap.Error(err))
		return
	}

	// Cache all transactions with 24-hour expiration
	cached := 0
	for _, transaction := range transactions {
		bf.cache.SetTransaction(transaction.ID, transaction, 24*time.Hour)
		cached++
	}

	duration := time.Since(startTime)
	logger.Info("Successfully fetched and cached transactions",
		zap.String("provider", providerName),
		zap.Int("count", len(transactions)),
		zap.Int("cached", cached),
		zap.Duration("duration", duration))
}
