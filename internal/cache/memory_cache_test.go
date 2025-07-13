package cache

import (
	"testing"
	"time"

	"github.com/rogerwesterbo/svennescamping-backend/internal/services/prices"
	"github.com/rogerwesterbo/svennescamping-backend/pkg/entities"
)

func TestInMemoryCache_Transactions(t *testing.T) {
	cache := NewInMemoryCache(1*time.Hour, 10*time.Minute)

	// Test transaction caching
	transaction := entities.Transaction{
		ID:         "test_transaction_1",
		ExternalID: "ext_123",
		Source:     "stripe",
		Amount:     199.99,
		Currency:   "NOK",
		Status:     "completed",
		CreatedAt:  time.Now(),
		CachedAt:   time.Now(),
	}

	// Set transaction
	cache.SetTransaction(transaction.ID, transaction, 1*time.Hour)

	// Get transaction
	retrieved, found := cache.GetTransaction(transaction.ID)
	if !found {
		t.Errorf("Expected to find transaction, but it was not found")
	}

	if retrieved.ID != transaction.ID {
		t.Errorf("Expected transaction ID %s, got %s", transaction.ID, retrieved.ID)
	}

	if retrieved.Amount != transaction.Amount {
		t.Errorf("Expected transaction amount %f, got %f", transaction.Amount, retrieved.Amount)
	}

	// Test getting all transactions
	transactions := cache.GetTransactions("")
	if len(transactions) != 1 {
		t.Errorf("Expected 1 transaction, got %d", len(transactions))
	}

	// Test pattern matching
	transactions = cache.GetTransactions("test_transaction")
	if len(transactions) != 1 {
		t.Errorf("Expected 1 transaction with pattern, got %d", len(transactions))
	}

	// Test pattern that doesn't match
	transactions = cache.GetTransactions("nonexistent")
	if len(transactions) != 0 {
		t.Errorf("Expected 0 transactions with non-matching pattern, got %d", len(transactions))
	}

	// Test deleting transaction
	cache.DeleteTransaction(transaction.ID)
	_, found = cache.GetTransaction(transaction.ID)
	if found {
		t.Errorf("Expected transaction to be deleted, but it was still found")
	}
}

func TestInMemoryCache_Prices(t *testing.T) {
	cache := NewInMemoryCache(1*time.Hour, 10*time.Minute)

	// Test price caching
	price := prices.Price{
		Product:  "test_product",
		Currency: "NOK",
		Price:    50.0,
	}

	// Set price (no expiration)
	cache.SetPrice(price.Product, price)

	// Get price
	retrieved, found := cache.GetPrice(price.Product)
	if !found {
		t.Errorf("Expected to find price, but it was not found")
	}

	if retrieved.Product != price.Product {
		t.Errorf("Expected price product %s, got %s", price.Product, retrieved.Product)
	}

	if retrieved.Price != price.Price {
		t.Errorf("Expected price %f, got %f", price.Price, retrieved.Price)
	}

	if retrieved.Currency != price.Currency {
		t.Errorf("Expected price currency %s, got %s", price.Currency, retrieved.Currency)
	}

	// Test getting all prices
	allPrices := cache.GetPrices()
	if len(allPrices) != 1 {
		t.Errorf("Expected 1 price, got %d", len(allPrices))
	}

	// Test deleting price
	cache.DeletePrice(price.Product)
	_, found = cache.GetPrice(price.Product)
	if found {
		t.Errorf("Expected price to be deleted, but it was still found")
	}
}

func TestInMemoryCache_MultipleTransactions(t *testing.T) {
	cache := NewInMemoryCache(1*time.Hour, 10*time.Minute)

	// Add multiple transactions
	transactions := []entities.Transaction{
		{
			ID:        "tx1",
			Source:    "stripe",
			Amount:    100.0,
			Currency:  "NOK",
			Status:    "completed",
			CreatedAt: time.Now(),
			CachedAt:  time.Now(),
		},
		{
			ID:        "tx2",
			Source:    "vipps",
			Amount:    200.0,
			Currency:  "NOK",
			Status:    "completed",
			CreatedAt: time.Now(),
			CachedAt:  time.Now(),
		},
		{
			ID:        "tx3",
			Source:    "zettle",
			Amount:    300.0,
			Currency:  "NOK",
			Status:    "completed",
			CreatedAt: time.Now(),
			CachedAt:  time.Now(),
		},
	}

	// Set all transactions
	for _, tx := range transactions {
		cache.SetTransaction(tx.ID, tx, 1*time.Hour)
	}

	// Get all transactions
	allTransactions := cache.GetTransactions("")
	if len(allTransactions) != 3 {
		t.Errorf("Expected 3 transactions, got %d", len(allTransactions))
	}

	// Verify each transaction can be retrieved individually
	for _, expected := range transactions {
		retrieved, found := cache.GetTransaction(expected.ID)
		if !found {
			t.Errorf("Expected to find transaction %s", expected.ID)
			continue
		}
		if retrieved.ID != expected.ID {
			t.Errorf("Expected transaction ID %s, got %s", expected.ID, retrieved.ID)
		}
		if retrieved.Amount != expected.Amount {
			t.Errorf("Expected transaction amount %f, got %f", expected.Amount, retrieved.Amount)
		}
	}
}

func TestInMemoryCache_Clear(t *testing.T) {
	cache := NewInMemoryCache(1*time.Hour, 10*time.Minute)

	// Add some data
	transaction := entities.Transaction{
		ID:        "test_tx",
		Source:    "stripe",
		Amount:    100.0,
		Currency:  "NOK",
		Status:    "completed",
		CreatedAt: time.Now(),
		CachedAt:  time.Now(),
	}

	price := prices.Price{
		Product:  "test_product",
		Currency: "NOK",
		Price:    50.0,
	}

	cache.SetTransaction(transaction.ID, transaction, 1*time.Hour)
	cache.SetPrice(price.Product, price)

	// Verify data is there
	_, foundTx := cache.GetTransaction(transaction.ID)
	_, foundPrice := cache.GetPrice(price.Product)
	if !foundTx || !foundPrice {
		t.Errorf("Expected to find both transaction and price before clear")
	}

	// Clear cache
	cache.Clear()

	// Verify data is gone
	_, foundTx = cache.GetTransaction(transaction.ID)
	_, foundPrice = cache.GetPrice(price.Product)
	if foundTx || foundPrice {
		t.Errorf("Expected cache to be empty after clear")
	}

	// Verify GetTransactions and GetPrices return empty slices
	transactions := cache.GetTransactions("")
	priceList := cache.GetPrices()
	if len(transactions) != 0 || len(priceList) != 0 {
		t.Errorf("Expected empty slices after clear, got %d transactions and %d prices", len(transactions), len(priceList))
	}
}

func TestInMemoryCache_Expiration(t *testing.T) {
	cache := NewInMemoryCache(1*time.Hour, 10*time.Minute)

	// Test transaction expiration
	transaction := entities.Transaction{
		ID:        "expiring_tx",
		Source:    "stripe",
		Amount:    100.0,
		Currency:  "NOK",
		Status:    "completed",
		CreatedAt: time.Now(),
		CachedAt:  time.Now(),
	}

	// Set transaction with very short expiration
	cache.SetTransaction(transaction.ID, transaction, 1*time.Millisecond)

	// Immediately check - should still be there
	_, found := cache.GetTransaction(transaction.ID)
	if !found {
		t.Errorf("Expected transaction to be found immediately after setting")
	}

	// Wait for expiration
	time.Sleep(10 * time.Millisecond)

	// Should be expired now
	_, found = cache.GetTransaction(transaction.ID)
	if found {
		t.Errorf("Expected transaction to be expired")
	}

	// Prices should not expire (NoExpiration)
	price := prices.Price{
		Product:  "persistent_product",
		Currency: "NOK",
		Price:    50.0,
	}

	cache.SetPrice(price.Product, price)

	// Even after waiting, price should still be there
	time.Sleep(10 * time.Millisecond)
	_, found = cache.GetPrice(price.Product)
	if !found {
		t.Errorf("Expected price to persist (no expiration)")
	}
}
