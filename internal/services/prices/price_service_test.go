package prices

import (
	"os"
	"path/filepath"
	"testing"
)

// createTestCSV creates a temporary CSV file for testing
func createTestCSV(t *testing.T) string {
	content := `Product;Price;Currency
Cabin;650;NOK
Bed linen;75;NOK
Caravan/motorhome/tent 1-2 pers;390;NOK
Caravan/motorhome/tent 3 pers;410;NOK
Caravan/motorhome/tent 4 pers;430;NOK`

	tmpDir := t.TempDir()
	csvPath := filepath.Join(tmpDir, "test_prices.csv")

	err := os.WriteFile(csvPath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test CSV file: %v", err)
	}

	return csvPath
}

func TestNewPriceService(t *testing.T) {
	csvPath := createTestCSV(t)

	service, err := NewPriceService(csvPath)
	if err != nil {
		t.Fatalf("Failed to create PriceService: %v", err)
	}

	if len(service.prices) != 5 {
		t.Errorf("Expected 5 prices, got %d", len(service.prices))
	}
}

func TestGetPriceByProduct(t *testing.T) {
	csvPath := createTestCSV(t)
	service, _ := NewPriceService(csvPath)

	tests := []struct {
		product       string
		expectedPrice float64
		expectError   bool
	}{
		{"Cabin", 650, false},
		{"Bed linen", 75, false},
		{"cabin", 650, false}, // Case insensitive
		{"NonExistent", 0, true},
	}

	for _, test := range tests {
		price, err := service.GetPriceByProduct(test.product)

		if test.expectError {
			if err == nil {
				t.Errorf("Expected error for product '%s', but got none", test.product)
			}
		} else {
			if err != nil {
				t.Errorf("Unexpected error for product '%s': %v", test.product, err)
			}
			if price.Price != test.expectedPrice {
				t.Errorf("Expected price %.2f for product '%s', got %.2f", test.expectedPrice, test.product, price.Price)
			}
		}
	}
}

func TestGetProductsByPrice(t *testing.T) {
	csvPath := createTestCSV(t)
	service, _ := NewPriceService(csvPath)

	// Test finding products with price 650
	products, err := service.GetProductsByPrice(650)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(products) != 1 {
		t.Errorf("Expected 1 product with price 650, got %d", len(products))
	}
	if products[0].Product != "Cabin" {
		t.Errorf("Expected product 'Cabin', got '%s'", products[0].Product)
	}

	// Test finding products with non-existent price
	_, err = service.GetProductsByPrice(999)
	if err == nil {
		t.Error("Expected error for non-existent price, but got none")
	}
}

func TestGetProductsByPriceRange(t *testing.T) {
	csvPath := createTestCSV(t)
	service, _ := NewPriceService(csvPath)

	// Test price range 400-450
	products, err := service.GetProductsByPriceRange(400, 450)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(products) != 2 {
		t.Errorf("Expected 2 products in range 400-450, got %d", len(products))
	}

	// Test invalid range
	_, err = service.GetProductsByPriceRange(500, 400)
	if err == nil {
		t.Error("Expected error for invalid price range, but got none")
	}
}

func TestGetAllPrices(t *testing.T) {
	csvPath := createTestCSV(t)
	service, _ := NewPriceService(csvPath)

	allPrices := service.GetAllPrices()
	if len(allPrices) != 5 {
		t.Errorf("Expected 5 prices, got %d", len(allPrices))
	}
}

func TestGetAllProducts(t *testing.T) {
	csvPath := createTestCSV(t)
	service, _ := NewPriceService(csvPath)

	allProducts := service.GetAllProducts()
	if len(allProducts) != 5 {
		t.Errorf("Expected 5 products, got %d", len(allProducts))
	}
}
