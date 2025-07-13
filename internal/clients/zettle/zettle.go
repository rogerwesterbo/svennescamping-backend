package zettle

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rogerwesterbo/svennescamping-backend/pkg/consts"
	"github.com/rogerwesterbo/svennescamping-backend/pkg/entities"
	"github.com/rogerwesterbo/svennescamping-backend/pkg/helpers/statushelpers"
	"github.com/rogerwesterbo/svennescamping-backend/pkg/interfaces"
	"github.com/rogerwesterbo/svennescamping-backend/pkg/logger"
	"go.uber.org/zap"
)

type ZettleClient struct {
	APIKey       string
	APIURL       string
	OAuthURL     string
	ClientID     string
	ClientSecret string
	httpClient   *http.Client

	// Token management
	accessToken string
	tokenExpiry time.Time
}

type ZettlePayment struct {
	UUID      string    `json:"uuid"`
	Amount    int64     `json:"amount"` // Amount in øre
	Currency  string    `json:"currency"`
	Timestamp time.Time `json:"timestamp"`
	Reference string    `json:"reference"`
	CardType  string    `json:"cardType"`
}

type ZettlePaymentsResponse struct {
	Purchases []ZettlePayment `json:"purchases"`
}

// Compile-time check to ensure ZettleClient implements Transactions interface
var _ interfaces.Transactions = (*ZettleClient)(nil)

func NewZettleClient(apiKey, apiURL, clientID, secret string) *ZettleClient {
	logger.Info("Initializing Zettle client",
		zap.String("api_url", apiURL),
		zap.String("client_id", clientID),
		zap.Bool("has_api_key", apiKey != ""),
		zap.Bool("has_secret", secret != ""))

	return &ZettleClient{
		APIKey:       apiKey,
		APIURL:       apiURL,
		OAuthURL:     "https://oauth.izettle.com", // Fixed OAuth URL
		ClientID:     clientID,
		ClientSecret: secret,
		httpClient:   &http.Client{Timeout: 30 * time.Second},
	}
}

func (z *ZettleClient) makeAuthenticatedRequest(ctx context.Context, method, requestURL string, body []byte) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		reqBody = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, requestURL, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Use API key authentication - most Zettle APIs use Bearer token with API key
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", z.APIKey))
	req.Header.Set("Content-Type", "application/json")

	resp, err := z.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return resp, nil
}

func (z *ZettleClient) GetLatestTransactions(ctx context.Context, limit int) ([]entities.Transaction, error) {
	logger.Info("Fetching transactions from Zettle", zap.Int("limit", limit))

	// Calculate date range for the last 30 days
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -30)

	// Format dates as required by Zettle API (YYYY-MM-DD)
	startDateStr := startDate.Format("2006-01-02")
	endDateStr := endDate.Format("2006-01-02")

	// Use correct Zettle Purchase API endpoint with required parameters
	// Documentation: https://developer.zettle.com/docs/api/purchase-retrieval
	endpoint := fmt.Sprintf("/purchases/v2?startDate=%s&endDate=%s", startDateStr, endDateStr)

	// Add limit parameter (Zettle typically supports up to 1000)
	if limit > 0 {
		if limit > 1000 {
			limit = 1000 // Zettle API limit
		}
		endpoint += fmt.Sprintf("&limit=%d", limit)
	}

	requestURL := fmt.Sprintf("%s%s", z.APIURL, endpoint)
	logger.Info("Making Zettle API request",
		zap.String("url", requestURL),
		zap.String("base_url", z.APIURL),
		zap.String("endpoint", endpoint),
		zap.String("date_range", fmt.Sprintf("%s to %s", startDateStr, endDateStr)))

	resp, err := z.makeAuthenticatedRequest(ctx, "GET", requestURL, nil)
	if err != nil {
		logger.Error("Failed to make Zettle API request", zap.Error(err))
		return nil, fmt.Errorf("failed to fetch transactions from Zettle: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		bodyString := string(bodyBytes)
		logger.Error("Zettle API returned error status",
			zap.Int("status", resp.StatusCode),
			zap.String("response_body", bodyString),
			zap.String("url", requestURL))

		// Handle specific Zettle error cases
		if resp.StatusCode == 401 {
			return nil, fmt.Errorf("Zettle authentication failed - check client credentials and access token")
		}
		if resp.StatusCode == 404 {
			return nil, fmt.Errorf("Zettle endpoint not found - check API URL: %s", z.APIURL)
		}

		return nil, fmt.Errorf("Zettle API returned status %d: %s", resp.StatusCode, bodyString)
	}

	var zettleResp ZettlePaymentsResponse
	if err := json.NewDecoder(resp.Body).Decode(&zettleResp); err != nil {
		logger.Error("Failed to decode Zettle response", zap.Error(err))
		return nil, fmt.Errorf("failed to decode Zettle response: %w", err)
	}

	var transactions []entities.Transaction
	count := 0

	for _, zp := range zettleResp.Purchases {
		if count >= limit {
			break
		}

		transaction := entities.Transaction{
			ID:              fmt.Sprintf("zettle_internal_%s", zp.UUID),
			ExternalID:      zp.UUID,
			Source:          consts.PAYMENT_SOURCE_ZETTLE,
			Amount:          float64(zp.Amount) / 100, // Convert from øre to NOK
			Currency:        zp.Currency,
			Status:          statushelpers.NormalizeTransactionStatus("COMPLETED", consts.PAYMENT_SOURCE_ZETTLE),
			CreatedAt:       zp.Timestamp,
			TransactionType: "card_payment",
			Description:     fmt.Sprintf("Zettle %s payment", zp.CardType),
			PaymentMethod:   "card",
			Metadata:        map[string]string{"provider": "zettle", "card_type": zp.CardType, "reference": zp.Reference},
			CachedAt:        time.Now(),
		}
		transactions = append(transactions, transaction)
		count++
	}

	logger.Info("Successfully fetched Zettle transactions", zap.Int("count", len(transactions)))
	return transactions, nil
}

func (z *ZettleClient) GetTransactionByID(ctx context.Context, id string) (entities.Transaction, error) {
	logger.Info("Fetching transaction by ID from Zettle", zap.String("id", id))

	// Use Zettle Payments API to get specific purchase details
	requestURL := fmt.Sprintf("%s/purchase/%s", z.APIURL, id)

	resp, err := z.makeAuthenticatedRequest(ctx, "GET", requestURL, nil)
	if err != nil {
		logger.Error("Failed to make Zettle API request", zap.Error(err), zap.String("id", id))
		return entities.Transaction{}, fmt.Errorf("failed to fetch transaction from Zettle: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return entities.Transaction{}, fmt.Errorf("transaction not found: %s", id)
	}

	if resp.StatusCode != http.StatusOK {
		logger.Error("Zettle API returned error status", zap.Int("status", resp.StatusCode), zap.String("id", id))
		return entities.Transaction{}, fmt.Errorf("Zettle API returned status %d", resp.StatusCode)
	}

	var zp ZettlePayment
	if err := json.NewDecoder(resp.Body).Decode(&zp); err != nil {
		logger.Error("Failed to decode Zettle response", zap.Error(err), zap.String("id", id))
		return entities.Transaction{}, fmt.Errorf("failed to decode Zettle response: %w", err)
	}

	transaction := entities.Transaction{
		ID:              fmt.Sprintf("zettle_internal_%s", zp.UUID),
		ExternalID:      zp.UUID,
		Source:          consts.PAYMENT_SOURCE_ZETTLE,
		Amount:          float64(zp.Amount) / 100, // Convert from øre to NOK
		Currency:        zp.Currency,
		Status:          statushelpers.NormalizeTransactionStatus("COMPLETED", consts.PAYMENT_SOURCE_ZETTLE),
		CreatedAt:       zp.Timestamp,
		TransactionType: "card_payment",
		Description:     fmt.Sprintf("Zettle %s payment", zp.CardType),
		PaymentMethod:   "card",
		Metadata:        map[string]string{"provider": "zettle", "card_type": zp.CardType, "reference": zp.Reference},
		CachedAt:        time.Now(),
	}

	return transaction, nil
}
