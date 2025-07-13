package interfaces

import (
	"context"

	"github.com/rogerwesterbo/svennescamping-backend/pkg/entities"
)

type TransactionRepository interface {
	GetTransactions(ctx context.Context, limit int) ([]entities.Transaction, error)
	GetTransactionByID(ctx context.Context, id string) (entities.Transaction, error)
	RefreshCache(ctx context.Context) error
}
