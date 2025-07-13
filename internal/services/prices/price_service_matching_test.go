package prices

import (
	"testing"
)

func TestPriceService_FindBestProductMatch(t *testing.T) {
	// Create a test price service with sample data
	ps := &PriceService{
		prices: []Price{
			{Product: "Cabin", Price: 650.0, Currency: "NOK"},
			{Product: "Caravan/motorhome/tent 1-2 pers", Price: 390.0, Currency: "NOK"},
			{Product: "Caravan/motorhome/tent 3 pers", Price: 410.0, Currency: "NOK"},
			{Product: "Caravan/motorhome/tent 4 pers", Price: 430.0, Currency: "NOK"},
			{Product: "Washing machine", Price: 40.0, Currency: "NOK"},
			{Product: "Bed linen", Price: 75.0, Currency: "NOK"},
			{Product: "Shower", Price: 15.0, Currency: "NOK"},
		},
	}

	tests := []struct {
		name        string
		amount      float64
		description string
		wantProduct *string
		wantPrice   *float64
	}{
		{
			name:        "Exact price match - Cabin",
			amount:      650.0,
			description: "",
			wantProduct: stringPtr("Cabin"),
			wantPrice:   float64Ptr(650.0),
		},
		{
			name:        "Exact price match with description confirmation",
			amount:      390.0,
			description: "caravan booking",
			wantProduct: stringPtr("Caravan/motorhome/tent 1-2 pers"),
			wantPrice:   float64Ptr(390.0),
		},
		{
			name:        "Description match with different price",
			amount:      400.0, // Close to 390 but not exact
			description: "tent 1-2 pers",
			wantProduct: stringPtr("Caravan/motorhome/tent 1-2 pers"),
			wantPrice:   float64Ptr(390.0),
		},
		{
			name:        "Price range match - shower",
			amount:      14.5, // Close to 15.0
			description: "",
			wantProduct: stringPtr("Shower"),
			wantPrice:   float64Ptr(15.0),
		},
		{
			name:        "Multiple price matches, description disambiguates",
			amount:      410.0,
			description: "tent 3 pers booking",
			wantProduct: stringPtr("Caravan/motorhome/tent 3 pers"),
			wantPrice:   float64Ptr(410.0),
		},
		{
			name:        "No match found",
			amount:      999.0,
			description: "unknown service",
			wantProduct: nil,
			wantPrice:   nil,
		},
		{
			name:        "Washing machine exact match",
			amount:      40.0,
			description: "laundry",
			wantProduct: stringPtr("Washing machine"),
			wantPrice:   float64Ptr(40.0),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := ps.FindBestProductMatch(tc.amount, tc.description)

			if tc.wantProduct == nil && tc.wantPrice == nil {
				if result != nil {
					t.Errorf("FindBestProductMatch() = %v, want nil", result)
				}
				return
			}

			if result == nil {
				t.Errorf("FindBestProductMatch() = nil, want product=%s, price=%f", *tc.wantProduct, *tc.wantPrice)
				return
			}

			if result.Product != *tc.wantProduct {
				t.Errorf("FindBestProductMatch() product = %v, want %v", result.Product, *tc.wantProduct)
			}

			if result.Price != *tc.wantPrice {
				t.Errorf("FindBestProductMatch() price = %v, want %v", result.Price, *tc.wantPrice)
			}
		})
	}
}

func TestPriceService_fuzzyMatch(t *testing.T) {
	ps := &PriceService{}

	tests := []struct {
		description string
		productName string
		want        bool
	}{
		{"cabin", "Cabin", true},
		{"CABIN BOOKING", "Cabin", true},
		{"tent 2 pers", "Caravan/motorhome/tent 1-2 pers", true},
		{"caravan for 4 people", "Caravan/motorhome/tent 4 pers", true},
		{"washing", "Washing machine", true},
		{"laundry service", "Washing machine", false}, // "laundry" doesn't contain "washing"
		{"shower time", "Shower", true},
		{"completely different", "Cabin", false},
		{"", "Cabin", false},
		{"bed linen rental", "Bed linen", true},
		{"motorhome", "Caravan/motorhome/tent 4 pers", true},
		{"tent for 3", "Caravan/motorhome/tent 3 pers", true},
	}

	for _, tc := range tests {
		t.Run(tc.description+"->"+tc.productName, func(t *testing.T) {
			if got := ps.fuzzyMatch(tc.description, tc.productName); got != tc.want {
				t.Errorf("fuzzyMatch(%q, %q) = %v, want %v", tc.description, tc.productName, got, tc.want)
			}
		})
	}
}

// Helper functions for creating pointers
func stringPtr(s string) *string {
	return &s
}

func float64Ptr(f float64) *float64 {
	return &f
}
