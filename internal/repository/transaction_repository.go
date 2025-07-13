package repository

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/rogerwesterbo/svennescamping-backend/pkg/consts"
	"github.com/rogerwesterbo/svennescamping-backend/pkg/entities"
	"github.com/rogerwesterbo/svennescamping-backend/pkg/interfaces"
	"github.com/rogerwesterbo/svennescamping-backend/pkg/logger"
	"go.uber.org/zap"
)

type TransactionRepository struct {
	cache        interfaces.Cache
	stripeClient interfaces.Transactions
	vippsClient  interfaces.Transactions
	zettleClient interfaces.Transactions
}

// Compile-time check to ensure TransactionRepository implements TransactionRepository interface
var _ interfaces.TransactionRepository = (*TransactionRepository)(nil)

func NewTransactionRepository(
	cache interfaces.Cache,
	stripeClient interfaces.Transactions,
	vippsClient interfaces.Transactions,
	zettleClient interfaces.Transactions,
) *TransactionRepository {
	return &TransactionRepository{
		cache:        cache,
		stripeClient: stripeClient,
		vippsClient:  vippsClient,
		zettleClient: zettleClient,
	}
}

func (r *TransactionRepository) GetTransactions(ctx context.Context, limit int) ([]entities.Transaction, error) {
	// Validate limit
	if limit < consts.TRANSACTION_LIMIT_MIN {
		limit = consts.TRANSACTION_LIMIT_DEFAULT
	}
	if limit > consts.TRANSACTION_LIMIT_MAX {
		limit = consts.TRANSACTION_LIMIT_MAX
	}

	// Get all cached transactions
	cachedTransactions := r.cache.GetTransactions("")

	// Sort by CreatedAt descending (newest first)
	sort.Slice(cachedTransactions, func(i, j int) bool {
		return cachedTransactions[i].CreatedAt.After(cachedTransactions[j].CreatedAt)
	})

	// Return the requested limit from cache
	if len(cachedTransactions) >= limit {
		return cachedTransactions[:limit], nil
	}

	// If we don't have enough cached data, return what we have
	// The background fetcher should be populating the cache automatically
	if len(cachedTransactions) > 0 {
		logger.Info("Returning partial transaction data from cache",
			zap.Int("available", len(cachedTransactions)),
			zap.Int("requested", limit))
		return cachedTransactions, nil
	}

	// If cache is completely empty, try to refresh once as fallback
	logger.Warn("Cache is empty, performing one-time refresh as fallback")
	err := r.RefreshCache(ctx)
	if err != nil {
		logger.Error("Failed to refresh cache as fallback", zap.Error(err))
		return []entities.Transaction{}, fmt.Errorf("no transactions available and failed to refresh cache: %w", err)
	}

	// Get updated cached transactions after refresh
	cachedTransactions = r.cache.GetTransactions("")

	// Sort by CreatedAt descending (newest first)
	sort.Slice(cachedTransactions, func(i, j int) bool {
		return cachedTransactions[i].CreatedAt.After(cachedTransactions[j].CreatedAt)
	})

	// Return the requested limit
	if len(cachedTransactions) > limit {
		return cachedTransactions[:limit], nil
	}
	return cachedTransactions, nil
}

func (r *TransactionRepository) GetTransactionByID(ctx context.Context, id string) (entities.Transaction, error) {
	// First check cache
	if transaction, found := r.cache.GetTransaction(id); found {
		return transaction, nil
	}

	// If not in cache, try to find it from each provider
	// Try Stripe first (check if it looks like a Stripe ID)
	if r.stripeClient != nil {
		transaction, err := r.stripeClient.GetTransactionByID(ctx, id)
		if err == nil {
			// Cache the transaction
			r.cache.SetTransaction(transaction.ID, transaction, 24*time.Hour)
			return transaction, nil
		}
		logger.Debug("Transaction not found in Stripe", zap.String("id", id), zap.Error(err))
	}

	// Try Vipps
	if r.vippsClient != nil {
		transaction, err := r.vippsClient.GetTransactionByID(ctx, id)
		if err == nil {
			// Cache the transaction
			r.cache.SetTransaction(transaction.ID, transaction, 24*time.Hour)
			return transaction, nil
		}
		logger.Debug("Transaction not found in Vipps", zap.String("id", id), zap.Error(err))
	}

	// Try Zettle
	if r.zettleClient != nil {
		transaction, err := r.zettleClient.GetTransactionByID(ctx, id)
		if err == nil {
			// Cache the transaction
			r.cache.SetTransaction(transaction.ID, transaction, 24*time.Hour)
			return transaction, nil
		}
		logger.Debug("Transaction not found in Zettle", zap.String("id", id), zap.Error(err))
	}

	return entities.Transaction{}, fmt.Errorf("transaction with ID %s not found", id)
}

func (r *TransactionRepository) RefreshCache(ctx context.Context) error {
	var allTransactions []entities.Transaction

	// Fetch from Stripe
	if r.stripeClient != nil {
		stripeTransactions, err := r.stripeClient.GetLatestTransactions(ctx, 100) // Fetch more for cache
		if err != nil {
			logger.Error("Failed to fetch Stripe transactions", zap.Error(err))
		} else {
			allTransactions = append(allTransactions, stripeTransactions...)
			logger.Info("Fetched Stripe transactions", zap.Int("count", len(stripeTransactions)))
		}
	}

	// Fetch from Vipps
	if r.vippsClient != nil {
		vippsTransactions, err := r.vippsClient.GetLatestTransactions(ctx, 100)
		if err != nil {
			logger.Error("Failed to fetch Vipps transactions", zap.Error(err))
		} else {
			allTransactions = append(allTransactions, vippsTransactions...)
			logger.Info("Fetched Vipps transactions", zap.Int("count", len(vippsTransactions)))
		}
	}

	// Fetch from Zettle
	if r.zettleClient != nil {
		zettleTransactions, err := r.zettleClient.GetLatestTransactions(ctx, 100)
		if err != nil {
			logger.Error("Failed to fetch Zettle transactions", zap.Error(err))
		} else {
			allTransactions = append(allTransactions, zettleTransactions...)
			logger.Info("Fetched Zettle transactions", zap.Int("count", len(zettleTransactions)))
		}
	}

	// Cache all transactions with 24-hour expiration
	for _, transaction := range allTransactions {
		r.cache.SetTransaction(transaction.ID, transaction, 24*time.Hour)
	}

	logger.Info("Refreshed transaction cache", zap.Int("total_transactions", len(allTransactions)))
	return nil
}
