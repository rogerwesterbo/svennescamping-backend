package vipps

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestVippsClient_getAccessToken(t *testing.T) {
	// Create a mock server to simulate Vipps token endpoint
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/accesstoken/get" {
			t.Errorf("Expected path /accesstoken/get, got %s", r.URL.Path)
		}

		if r.Method != "POST" {
			t.Errorf("Expected POST method, got %s", r.Method)
		}

		// Check required headers
		if r.Header.Get("client_id") == "" {
			t.Error("Missing client_id header")
		}
		if r.Header.Get("client_secret") == "" {
			t.Error("Missing client_secret header")
		}
		if r.Header.Get("Ocp-Apim-Subscription-Key") == "" {
			t.Error("Missing Ocp-Apim-Subscription-Key header")
		}

		// Return mock token response
		response := TokenResponse{
			TokenType:   "Bearer",
			ExpiresIn:   "3600",
			AccessToken: "mock_access_token_12345",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer mockServer.Close()

	// Create Vipps client with mock server URL
	client := NewVippsClient("test_subscription_key", mockServer.URL, "test_client_id", "test_secret", "123456")

	ctx := context.Background()

	// Test getting access token
	token, err := client.getAccessToken(ctx)
	if err != nil {
		t.Fatalf("Failed to get access token: %v", err)
	}

	if token != "mock_access_token_12345" {
		t.Errorf("Expected token 'mock_access_token_12345', got '%s'", token)
	}

	// Test token caching - should return the same token without making another request
	token2, err := client.getAccessToken(ctx)
	if err != nil {
		t.Fatalf("Failed to get cached access token: %v", err)
	}

	if token2 != token {
		t.Errorf("Expected cached token to be the same, got different token")
	}
}

func TestVippsClient_tokenExpiry(t *testing.T) {
	callCount := 0

	// Create a mock server that tracks calls
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++

		// Return mock token response with short expiry
		response := TokenResponse{
			TokenType:   "Bearer",
			ExpiresIn:   "1", // 1 second expiry for testing
			AccessToken: "mock_access_token_" + string(rune(callCount+'0')),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer mockServer.Close()

	client := NewVippsClient("test_api_key", mockServer.URL, "test_client_id", "test_secret", "123456")

	ctx := context.Background()

	// Get first token
	token1, err := client.getAccessToken(ctx)
	if err != nil {
		t.Fatalf("Failed to get first access token: %v", err)
	}

	if callCount != 1 {
		t.Errorf("Expected 1 API call, got %d", callCount)
	}

	// Wait for token to expire
	time.Sleep(2 * time.Second)

	// Get second token - should make a new API call
	token2, err := client.getAccessToken(ctx)
	if err != nil {
		t.Fatalf("Failed to get second access token: %v", err)
	}

	if callCount != 2 {
		t.Errorf("Expected 2 API calls after expiry, got %d", callCount)
	}

	if token1 == token2 {
		t.Error("Expected different tokens after expiry, got the same token")
	}
}

func TestVippsClient_GetLatestTransactions(t *testing.T) {
	// Create a mock server for token and transactions endpoints
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/accesstoken/get" {
			response := TokenResponse{
				TokenType:   "Bearer",
				ExpiresIn:   "3600",
				AccessToken: "mock_access_token_12345",
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		} else if r.URL.Path == "/ecomm/v2/payments" {
			// Mock transactions response
			mockTransactions := VippsTransactionResponse{
				Transactions: []VippsTransaction{
					{
						TransactionID:   "vipps_tx_1",
						Amount:          15000, // 150.00 NOK in øre
						Currency:        "NOK",
						Status:          "completed",
						TimeStamp:       time.Now(),
						TransactionText: "Test payment 1",
						OrderID:         "order_1",
					},
					{
						TransactionID:   "vipps_tx_2",
						Amount:          25000, // 250.00 NOK in øre
						Currency:        "NOK",
						Status:          "completed",
						TimeStamp:       time.Now().Add(-time.Hour),
						TransactionText: "Test payment 2",
						OrderID:         "order_2",
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(mockTransactions)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer mockServer.Close()

	client := NewVippsClient("test_subscription_key", mockServer.URL, "test_client_id", "test_secret", "123456")

	ctx := context.Background()

	// Test getting transactions
	transactions, err := client.GetLatestTransactions(ctx, 3)
	if err != nil {
		t.Fatalf("Failed to get transactions: %v", err)
	}

	if len(transactions) != 2 {
		t.Errorf("Expected 2 transactions, got %d", len(transactions))
	}

	// Check the first transaction
	if len(transactions) > 0 {
		tx := transactions[0]
		if tx.Source != "vipps" {
			t.Errorf("Expected source 'vipps', got '%s'", tx.Source)
		}
		if tx.ExternalID != "vipps_tx_1" {
			t.Errorf("Expected ExternalID 'vipps_tx_1', got '%s'", tx.ExternalID)
		}
		if tx.Amount != 150.0 {
			t.Errorf("Expected amount 150.0, got %f", tx.Amount)
		}
		if tx.Currency != "NOK" {
			t.Errorf("Expected currency 'NOK', got '%s'", tx.Currency)
		}
		if tx.Metadata["order_id"] != "order_1" {
			t.Errorf("Expected order_id 'order_1', got '%s'", tx.Metadata["order_id"])
		}
	}
}
