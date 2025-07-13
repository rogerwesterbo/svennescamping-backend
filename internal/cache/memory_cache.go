package cache

import (
	"fmt"
	"strings"
	"time"

	gocache "github.com/patrickmn/go-cache"
	"github.com/rogerwesterbo/svennescamping-backend/internal/services/prices"
	"github.com/rogerwesterbo/svennescamping-backend/pkg/entities"
	"github.com/rogerwesterbo/svennescamping-backend/pkg/interfaces"
)

type InMemoryCache struct {
	cache *gocache.Cache
}

// Compile-time check to ensure InMemoryCache implements Cache interface
var _ interfaces.Cache = (*InMemoryCache)(nil)

func NewInMemoryCache(defaultExpiration, cleanupInterval time.Duration) *InMemoryCache {
	return &InMemoryCache{
		cache: gocache.New(defaultExpiration, cleanupInterval),
	}
}

// Transaction cache methods
func (c *InMemoryCache) SetTransaction(key string, transaction entities.Transaction, expiration time.Duration) {
	transactionKey := fmt.Sprintf("transaction:%s", key)
	c.cache.Set(transactionKey, transaction, expiration)
}

func (c *InMemoryCache) GetTransaction(key string) (entities.Transaction, bool) {
	transactionKey := fmt.Sprintf("transaction:%s", key)
	if item, found := c.cache.Get(transactionKey); found {
		if transaction, ok := item.(entities.Transaction); ok {
			return transaction, true
		}
	}
	return entities.Transaction{}, false
}

func (c *InMemoryCache) GetTransactions(pattern string) []entities.Transaction {
	var transactions []entities.Transaction
	items := c.cache.Items()

	for key, item := range items {
		if strings.HasPrefix(key, "transaction:") {
			if pattern == "" || strings.Contains(key, pattern) {
				if transaction, ok := item.Object.(entities.Transaction); ok {
					transactions = append(transactions, transaction)
				}
			}
		}
	}

	return transactions
}

func (c *InMemoryCache) DeleteTransaction(key string) {
	transactionKey := fmt.Sprintf("transaction:%s", key)
	c.cache.Delete(transactionKey)
}

// Price cache methods (no expiration)
func (c *InMemoryCache) SetPrice(key string, price prices.Price) {
	priceKey := fmt.Sprintf("price:%s", key)
	c.cache.Set(priceKey, price, gocache.NoExpiration)
}

func (c *InMemoryCache) GetPrice(key string) (prices.Price, bool) {
	priceKey := fmt.Sprintf("price:%s", key)
	if item, found := c.cache.Get(priceKey); found {
		if price, ok := item.(prices.Price); ok {
			return price, true
		}
	}
	return prices.Price{}, false
}

func (c *InMemoryCache) GetPrices() []prices.Price {
	var priceList []prices.Price
	items := c.cache.Items()

	for key, item := range items {
		if strings.HasPrefix(key, "price:") {
			if price, ok := item.Object.(prices.Price); ok {
				priceList = append(priceList, price)
			}
		}
	}

	return priceList
}

func (c *InMemoryCache) DeletePrice(key string) {
	priceKey := fmt.Sprintf("price:%s", key)
	c.cache.Delete(priceKey)
}

// Clear cache
func (c *InMemoryCache) Clear() {
	c.cache.Flush()
}
