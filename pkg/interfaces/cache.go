package interfaces

import (
	"time"

	"github.com/rogerwesterbo/svennescamping-backend/internal/services/prices"
	"github.com/rogerwesterbo/svennescamping-backend/pkg/entities"
)

type Cache interface {
	// Transaction cache methods
	SetTransaction(key string, transaction entities.Transaction, expiration time.Duration)
	GetTransaction(key string) (entities.Transaction, bool)
	GetTransactions(pattern string) []entities.Transaction
	DeleteTransaction(key string)

	// Price cache methods (no expiration)
	SetPrice(key string, price prices.Price)
	GetPrice(key string) (prices.Price, bool)
	GetPrices() []prices.Price
	DeletePrice(key string)

	// Clear cache
	Clear()
}
