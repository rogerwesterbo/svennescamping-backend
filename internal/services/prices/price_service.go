package prices

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Price represents a price entry from the CSV
type Price struct {
	Product  string
	Price    float64
	Currency string
}

// PriceService handles price-related operations
type PriceService struct {
	prices []Price
}

// NewPriceService creates a new PriceService and loads prices from the CSV file
func NewPriceService(csvFilePath string) (*PriceService, error) {
	service := &PriceService{
		prices: make([]Price, 0),
	}

	err := service.loadPricesFromCSV(csvFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load prices from CSV: %w", err)
	}

	return service, nil
}

// loadPricesFromCSV reads and parses the CSV file
func (ps *PriceService) loadPricesFromCSV(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = ';' // CSV uses semicolon as delimiter

	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to read CSV records: %w", err)
	}

	// Skip header row
	if len(records) < 2 {
		return fmt.Errorf("CSV file must contain at least header and one data row")
	}

	for i, record := range records[1:] { // Skip header
		if len(record) != 3 {
			return fmt.Errorf("invalid record at line %d: expected 3 columns, got %d", i+2, len(record))
		}

		price, err := strconv.ParseFloat(record[1], 64)
		if err != nil {
			return fmt.Errorf("invalid price value at line %d: %w", i+2, err)
		}

		ps.prices = append(ps.prices, Price{
			Product:  strings.TrimSpace(record[0]),
			Price:    price,
			Currency: strings.TrimSpace(record[2]),
		})
	}

	return nil
}

// GetPriceByProduct returns the price for a given product
func (ps *PriceService) GetPriceByProduct(product string) (*Price, error) {
	product = strings.TrimSpace(product)

	for _, p := range ps.prices {
		if strings.EqualFold(p.Product, product) {
			return &p, nil
		}
	}

	return nil, fmt.Errorf("product '%s' not found", product)
}

// GetProductsByPrice returns all products with the specified price
func (ps *PriceService) GetProductsByPrice(price float64) ([]Price, error) {
	var matchingProducts []Price

	for _, p := range ps.prices {
		if p.Price == price {
			matchingProducts = append(matchingProducts, p)
		}
	}

	if len(matchingProducts) == 0 {
		return nil, fmt.Errorf("no products found with price %.2f", price)
	}

	return matchingProducts, nil
}

// GetProductsByPriceRange returns all products within a price range (inclusive)
func (ps *PriceService) GetProductsByPriceRange(minPrice, maxPrice float64) ([]Price, error) {
	if minPrice > maxPrice {
		return nil, fmt.Errorf("minimum price cannot be greater than maximum price")
	}

	var matchingProducts []Price

	for _, p := range ps.prices {
		if p.Price >= minPrice && p.Price <= maxPrice {
			matchingProducts = append(matchingProducts, p)
		}
	}

	return matchingProducts, nil
}

// FindBestProductMatch attempts to find the best product match for a transaction
// It tries multiple strategies: exact price match, price range match, and description matching
func (ps *PriceService) FindBestProductMatch(amount float64, description string) *Price {
	// Strategy 1: Try exact price match first
	products, err := ps.GetProductsByPrice(amount)
	if err == nil && len(products) == 1 {
		// If exactly one product matches the price, it's likely correct
		return &products[0]
	}

	// Strategy 2: Try fuzzy description matching if we have a description
	if description != "" {
		for _, p := range ps.prices {
			if ps.fuzzyMatch(description, p.Product) {
				return &p
			}
		}
	}

	// Strategy 3: If multiple products match the price, try to disambiguate with description
	if err == nil && len(products) > 1 && description != "" {
		for _, p := range products {
			if ps.fuzzyMatch(description, p.Product) {
				return &p
			}
		}
		// If no description match but multiple price matches, return the first one
		return &products[0]
	}

	// Strategy 4: Try price range matching (±5% tolerance)
	tolerance := amount * 0.05
	rangeProducts, err := ps.GetProductsByPriceRange(amount-tolerance, amount+tolerance)
	if err == nil && len(rangeProducts) > 0 {
		// If we have a description, try to match within the range
		if description != "" {
			for _, p := range rangeProducts {
				if ps.fuzzyMatch(description, p.Product) {
					return &p
				}
			}
		}
		// Return the closest price match
		var closest *Price
		var minDiff float64 = tolerance + 1
		for _, p := range rangeProducts {
			diff := amount - p.Price
			if diff < 0 {
				diff = -diff
			}
			if diff < minDiff {
				minDiff = diff
				closest = &p
			}
		}
		if closest != nil {
			return closest
		}
	}

	return nil
}

// fuzzyMatch performs fuzzy string matching between description and product name
func (ps *PriceService) fuzzyMatch(description, productName string) bool {
	desc := strings.ToLower(strings.TrimSpace(description))
	prod := strings.ToLower(strings.TrimSpace(productName))

	// Handle empty strings
	if desc == "" || prod == "" {
		return false
	}

	// Exact match
	if desc == prod {
		return true
	}

	// Contains match
	if strings.Contains(desc, prod) || strings.Contains(prod, desc) {
		return true
	}

	// Split and check for word matches
	descWords := strings.Fields(desc)
	prodWords := strings.Fields(prod)

	if len(prodWords) == 0 {
		return false
	}

	matchCount := 0
	for _, dWord := range descWords {
		for _, pWord := range prodWords {
			// Clean up words (remove punctuation)
			cleanPWord := strings.Trim(pWord, "/-,.")
			cleanDWord := strings.Trim(dWord, "/-,.")

			if len(cleanPWord) > 2 && (strings.Contains(cleanDWord, cleanPWord) || strings.Contains(cleanPWord, cleanDWord)) {
				matchCount++
				break
			}

			// Special cases for numbers and common abbreviations
			if cleanPWord == "1-2" && (cleanDWord == "2" || cleanDWord == "1") {
				matchCount++
				break
			}
			if cleanPWord == "3" && (cleanDWord == "3" || cleanDWord == "three") {
				matchCount++
				break
			}
			if cleanPWord == "4" && (cleanDWord == "4" || cleanDWord == "four") {
				matchCount++
				break
			}
			if cleanPWord == "pers" && (cleanDWord == "people" || cleanDWord == "person") {
				matchCount++
				break
			}
		}
	}

	// Consider it a match if at least half the product words are found
	threshold := float64(len(prodWords)) * 0.4 // Lower threshold for better matching
	return float64(matchCount) >= threshold
}

// GetAllPrices returns all loaded prices
func (ps *PriceService) GetAllPrices() []Price {
	return ps.prices
}

// GetAllProducts returns all product names
func (ps *PriceService) GetAllProducts() []string {
	products := make([]string, len(ps.prices))
	for i, p := range ps.prices {
		products[i] = p.Product
	}
	return products
}
