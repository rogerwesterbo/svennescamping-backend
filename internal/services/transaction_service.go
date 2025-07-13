package services

import (
	"context"

	"github.com/rogerwesterbo/svennescamping-backend/pkg/entities"
	"github.com/rogerwesterbo/svennescamping-backend/pkg/interfaces"
	"github.com/rogerwesterbo/svennescamping-backend/pkg/logger"
	"go.uber.org/zap"
)

type TransactionService struct {
	repository interfaces.TransactionRepository
}

func NewTransactionService(repository interfaces.TransactionRepository) *TransactionService {
	return &TransactionService{
		repository: repository,
	}
}

func (s *TransactionService) GetTransactions(ctx context.Context, limit int) ([]entities.Transaction, error) {
	transactions, err := s.repository.GetTransactions(ctx, limit)
	if err != nil {
		return nil, err
	}

	// Enrich transactions with product information
	enrichedTransactions := s.enrichTransactionsWithProducts(transactions)
	return enrichedTransactions, nil
}

func (s *TransactionService) GetTransactionByID(ctx context.Context, id string) (entities.Transaction, error) {
	transaction, err := s.repository.GetTransactionByID(ctx, id)
	if err != nil {
		return entities.Transaction{}, err
	}

	// Enrich single transaction with product information
	enrichedTransaction := s.enrichTransactionWithProduct(transaction)
	return enrichedTransaction, nil
}

func (s *TransactionService) RefreshCache(ctx context.Context) error {
	return s.repository.RefreshCache(ctx)
}

// enrichTransactionsWithProducts enriches a slice of transactions with product information from the price list
func (s *TransactionService) enrichTransactionsWithProducts(transactions []entities.Transaction) []entities.Transaction {
	if PriceService == nil {
		logger.Warn("PriceService not available, skipping product enrichment")
		return transactions
	}

	enrichedTransactions := make([]entities.Transaction, len(transactions))
	for i, transaction := range transactions {
		enrichedTransactions[i] = s.enrichTransactionWithProduct(transaction)
	}

	return enrichedTransactions
}

// enrichTransactionWithProduct enriches a single transaction with product information from the price list
func (s *TransactionService) enrichTransactionWithProduct(transaction entities.Transaction) entities.Transaction {
	if PriceService == nil {
		return transaction
	}

	// Try to find a matching product
	matchedProduct := PriceService.FindBestProductMatch(transaction.Amount, transaction.Description)
	if matchedProduct != nil {
		// Create copies to avoid modifying the original transaction
		enrichedTransaction := transaction
		enrichedTransaction.Product = &matchedProduct.Product
		enrichedTransaction.ProductPrice = &matchedProduct.Price

		logger.Debug("Enriched transaction with product information",
			zap.String("transaction_id", transaction.ID),
			zap.String("matched_product", matchedProduct.Product),
			zap.Float64("product_price", matchedProduct.Price),
			zap.Float64("transaction_amount", transaction.Amount),
			zap.String("description", transaction.Description))

		return enrichedTransaction
	}

	logger.Debug("No product match found for transaction",
		zap.String("transaction_id", transaction.ID),
		zap.Float64("amount", transaction.Amount),
		zap.String("description", transaction.Description))

	return transaction
}
