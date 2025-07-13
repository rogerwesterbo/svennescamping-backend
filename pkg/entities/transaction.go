package entities

import "time"

type Transaction struct {
	ID              string            `json:"id"`
	ExternalID      string            `json:"external_id"` // ID from external provider
	Source          string            `json:"source"`      // Payment source: stripe, vipps, zettle
	Amount          float64           `json:"amount"`
	Currency        string            `json:"currency"`
	Status          string            `json:"status"`
	CreatedAt       time.Time         `json:"created_at"`
	CustomerID      string            `json:"customer_id"`
	Description     string            `json:"description"`
	PaymentMethod   string            `json:"payment_method"`
	ReceiptURL      string            `json:"receipt_url"`
	Metadata        map[string]string `json:"metadata"`
	TransactionType string            `json:"transaction_type"`        // e.g., "card", "bank_transfer"
	Data            any               `json:"data,omitempty"`          // Additional data if needed
	TransferData    any               `json:"transfer_data,omitempty"` // Data related to transfer, if applicable
	CachedAt        time.Time         `json:"cached_at"`               // When the transaction was cached
	// Product information enriched from price list
	Product      *string  `json:"product,omitempty"`       // Matched product name from price list
	ProductPrice *float64 `json:"product_price,omitempty"` // Expected price for the product
}
