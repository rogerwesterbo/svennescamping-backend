package vipps

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/rogerwesterbo/svennescamping-backend/pkg/consts"
	"github.com/rogerwesterbo/svennescamping-backend/pkg/entities"
	"github.com/rogerwesterbo/svennescamping-backend/pkg/helpers/statushelpers"
	"github.com/rogerwesterbo/svennescamping-backend/pkg/interfaces"
	"github.com/rogerwesterbo/svennescamping-backend/pkg/logger"
	"go.uber.org/zap"
)

type VippsClient struct {
	SubscriptionKey      string
	APIURL               string
	ClientID             string
	Secret               string
	MerchantSerialNumber string // Added required field
	httpClient           *http.Client

	// Token management
	accessToken string
	tokenExpiry time.Time
	tokenMutex  sync.RWMutex
}

type TokenResponse struct {
	TokenType    string `json:"token_type"`
	ExpiresIn    string `json:"expires_in"`
	ExtExpiresIn string `json:"ext_expires_in"`
	ExpiresOn    string `json:"expires_on"`
	NotBefore    string `json:"not_before"`
	Resource     string `json:"resource"`
	AccessToken  string `json:"access_token"`
}

type VippsTransaction struct {
	TransactionID   string    `json:"transactionId"`
	Amount          int       `json:"amount"` // Amount in øre (1 NOK = 100 øre)
	Currency        string    `json:"currency"`
	Status          string    `json:"status"`
	TimeStamp       time.Time `json:"timeStamp"`
	TransactionText string    `json:"transactionText"`
	OrderID         string    `json:"orderId"`
}

type VippsTransactionResponse struct {
	Transactions []VippsTransaction `json:"transactions"`
}

// Compile-time check to ensure VippsClient implements Transactions interface
var _ interfaces.Transactions = (*VippsClient)(nil)

func NewVippsClient(subscriptionKey, apiURL, clientID, secret, merchantSerialNumber string) *VippsClient {
	logger.Info("Initializing Vipps client",
		zap.String("api_url", apiURL),
		zap.String("client_id", clientID),
		zap.String("merchant_serial_number", merchantSerialNumber),
		zap.Bool("has_subscription_key", subscriptionKey != ""),
		zap.Bool("has_secret", secret != ""))

	return &VippsClient{
		SubscriptionKey:      subscriptionKey,
		APIURL:               apiURL,
		ClientID:             clientID,
		Secret:               secret,
		MerchantSerialNumber: merchantSerialNumber,
		httpClient:           &http.Client{Timeout: 30 * time.Second},
	}
}

func (v *VippsClient) getAccessToken(ctx context.Context) (string, error) {
	v.tokenMutex.RLock()
	// Check if we have a valid token that doesn't expire in the next 5 minutes
	if v.accessToken != "" && time.Now().Add(5*time.Minute).Before(v.tokenExpiry) {
		token := v.accessToken
		v.tokenMutex.RUnlock()
		return token, nil
	}
	v.tokenMutex.RUnlock()

	v.tokenMutex.Lock()
	defer v.tokenMutex.Unlock()

	// Double-check after acquiring write lock
	if v.accessToken != "" && time.Now().Add(5*time.Minute).Before(v.tokenExpiry) {
		return v.accessToken, nil
	}

	logger.Info("Fetching new Vipps access token")

	// Prepare the request
	tokenURL := fmt.Sprintf("%s/accesstoken/get", v.APIURL)
	logger.Info("Making Vipps token request", zap.String("token_url", tokenURL))

	req, err := http.NewRequestWithContext(ctx, "POST", tokenURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create token request: %w", err)
	}

	// Set required headers for Vipps token request
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("client_id", v.ClientID)
	req.Header.Set("client_secret", v.Secret)
	req.Header.Set("Ocp-Apim-Subscription-Key", v.SubscriptionKey)
	req.Header.Set("grant_type", "client_credentials") // Added per documentation

	// Make the request
	resp, err := v.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make token request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		bodyString := string(bodyBytes)
		logger.Error("Vipps token request failed",
			zap.Int("status", resp.StatusCode),
			zap.String("response_body", bodyString),
			zap.String("token_url", tokenURL))
		return "", fmt.Errorf("token request failed with status %d: %s", resp.StatusCode, bodyString)
	}

	// Parse the response
	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", fmt.Errorf("failed to decode token response: %w", err)
	}

	// Parse expiry time
	expiresIn, err := strconv.Atoi(tokenResp.ExpiresIn)
	if err != nil {
		logger.Warn("Failed to parse expires_in, using default 1 hour", zap.Error(err))
		expiresIn = 3600 // Default to 1 hour
	}

	// Store the token and expiry
	v.accessToken = tokenResp.AccessToken
	v.tokenExpiry = time.Now().Add(time.Duration(expiresIn) * time.Second)

	logger.Info("Successfully obtained Vipps access token",
		zap.Time("expires_at", v.tokenExpiry),
		zap.Int("expires_in_seconds", expiresIn))

	return v.accessToken, nil
}

func (v *VippsClient) makeAuthenticatedRequest(ctx context.Context, method, url string, body []byte) (*http.Response, error) {
	token, err := v.getAccessToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	var reqBody io.Reader
	if body != nil {
		reqBody = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set required headers
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Ocp-Apim-Subscription-Key", v.SubscriptionKey)

	// Add Merchant-Serial-Number if available (required for most Vipps APIs)
	if v.MerchantSerialNumber != "" {
		req.Header.Set("Merchant-Serial-Number", v.MerchantSerialNumber)
	}

	resp, err := v.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	// If we get 401, the token might have expired, try to refresh once
	if resp.StatusCode == http.StatusUnauthorized {
		resp.Body.Close()

		logger.Info("Received 401, refreshing Vipps token and retrying")

		// Clear the current token to force refresh
		v.tokenMutex.Lock()
		v.accessToken = ""
		v.tokenExpiry = time.Time{}
		v.tokenMutex.Unlock()

		// Get a new token
		newToken, err := v.getAccessToken(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to refresh access token: %w", err)
		}

		// Retry the request with new token
		var retryReqBody io.Reader
		if body != nil {
			retryReqBody = bytes.NewReader(body)
		}

		retryReq, err := http.NewRequestWithContext(ctx, method, url, retryReqBody)
		if err != nil {
			return nil, fmt.Errorf("failed to create retry request: %w", err)
		}

		retryReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", newToken))
		retryReq.Header.Set("Content-Type", "application/json")
		retryReq.Header.Set("Ocp-Apim-Subscription-Key", v.SubscriptionKey)

		// Add Merchant-Serial-Number if available
		if v.MerchantSerialNumber != "" {
			retryReq.Header.Set("Merchant-Serial-Number", v.MerchantSerialNumber)
		}

		resp, err = v.httpClient.Do(retryReq)
		if err != nil {
			return nil, fmt.Errorf("retry request failed: %w", err)
		}
	}

	return resp, nil
}

func (v *VippsClient) GetLatestTransactions(ctx context.Context, limit int) ([]entities.Transaction, error) {
	logger.Info("Fetching transactions from Vipps", zap.Int("limit", limit))

	// Calculate date range for the last 30 days
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -30)

	// Format dates as required by Vipps API (YYYY-MM-DD)
	since := startDate.Format("2006-01-02")
	until := endDate.Format("2006-01-02")

	// Try different Vipps API endpoints based on the official documentation
	// Note: Vipps doesn't support direct transaction listing - you need specific order IDs
	// or use the Reports API for settlement data
	possibleEndpoints := []string{
		// Reports API (recommended for transaction history)
		fmt.Sprintf("/report/v1/transactions?from=%s&to=%s", since, until),
		fmt.Sprintf("/report/v1/settlements?from=%s&to=%s", since, until),

		// Recurring API (if using Vipps Recurring)
		"/recurring/v2/agreements?status=ACTIVE",

		// ePayment API (requires specific order IDs, but let's try)
		"/ecomm/v2/payments",
		fmt.Sprintf("/ecomm/v2/payments?since=%s&until=%s", since, until),

		// Checkout API (newer Vipps product)
		"/checkout/v3/sessions",
	}

	var lastErr error

	// Try each endpoint until we find one that works
	for i, endpoint := range possibleEndpoints {
		// Add limit parameter if it makes sense for this endpoint
		testEndpoint := endpoint
		if limit > 0 && limit <= 100 && (i == 0 || i == 1 || i == 3 || i == 4) {
			separator := "&"
			if !strings.Contains(testEndpoint, "?") {
				separator = "?"
			}
			testEndpoint += fmt.Sprintf("%slimit=%d", separator, limit)
		}

		url := fmt.Sprintf("%s%s", v.APIURL, testEndpoint)
		logger.Info("Trying Vipps API endpoint",
			zap.Int("attempt", i+1),
			zap.String("url", url),
			zap.String("endpoint", testEndpoint))

		resp, err := v.makeAuthenticatedRequest(ctx, "GET", url, nil)
		if err != nil {
			logger.Warn("Failed to make request to endpoint",
				zap.String("endpoint", testEndpoint),
				zap.Error(err))
			lastErr = err
			continue
		}
		defer resp.Body.Close()

		// If we get 404, try the next endpoint
		if resp.StatusCode == http.StatusNotFound {
			logger.Info("Endpoint not found, trying next", zap.String("endpoint", testEndpoint))
			lastErr = fmt.Errorf("endpoint not found: %s", testEndpoint)
			continue
		}

		// If we get other errors, log them but continue trying
		if resp.StatusCode != http.StatusOK {
			bodyBytes, _ := io.ReadAll(resp.Body)
			bodyString := string(bodyBytes)
			logger.Warn("Endpoint returned error, trying next",
				zap.String("endpoint", testEndpoint),
				zap.Int("status", resp.StatusCode),
				zap.String("response_body", bodyString))
			lastErr = fmt.Errorf("endpoint %s returned status %d: %s", testEndpoint, resp.StatusCode, bodyString)
			continue
		}

		// Success! Parse the response
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			logger.Error("Failed to read response body", zap.Error(err))
			lastErr = err
			continue
		}

		logger.Info("Successfully connected to Vipps endpoint",
			zap.String("endpoint", testEndpoint),
			zap.String("response_preview", string(bodyBytes[:minInt(200, len(bodyBytes))])))

		// Try to parse as different response formats
		transactions, parseErr := v.parseVippsResponse(bodyBytes, testEndpoint)
		if parseErr != nil {
			logger.Warn("Failed to parse response from endpoint",
				zap.String("endpoint", testEndpoint),
				zap.Error(parseErr))
			lastErr = parseErr
			continue
		}

		// Apply limit if we got more transactions than requested
		if limit > 0 && len(transactions) > limit {
			transactions = transactions[:limit]
		}

		logger.Info("Successfully fetched Vipps transactions",
			zap.String("endpoint", testEndpoint),
			zap.Int("count", len(transactions)))
		return transactions, nil
	}

	// If we get here, none of the endpoints worked
	logger.Error("All Vipps API endpoints failed", zap.Error(lastErr))
	return nil, fmt.Errorf("failed to fetch transactions from any Vipps endpoint. Last error: %w. "+
		"This suggests your Vipps setup uses a different API product. Check your Vipps developer dashboard for the correct API endpoints.", lastErr)
}

func (v *VippsClient) GetTransactionByID(ctx context.Context, id string) (entities.Transaction, error) {
	logger.Info("Fetching transaction by ID from Vipps", zap.String("id", id))

	// Use Vipps eComm API to get specific payment details
	url := fmt.Sprintf("%s/ecomm/v2/payments/%s/details", v.APIURL, id)

	resp, err := v.makeAuthenticatedRequest(ctx, "GET", url, nil)
	if err != nil {
		logger.Error("Failed to make Vipps API request", zap.Error(err), zap.String("id", id))
		return entities.Transaction{}, fmt.Errorf("failed to fetch transaction from Vipps: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return entities.Transaction{}, fmt.Errorf("transaction not found: %s", id)
	}

	if resp.StatusCode != http.StatusOK {
		logger.Error("Vipps API returned error status", zap.Int("status", resp.StatusCode), zap.String("id", id))
		return entities.Transaction{}, fmt.Errorf("Vipps API returned status %d", resp.StatusCode)
	}

	var vt VippsTransaction
	if err := json.NewDecoder(resp.Body).Decode(&vt); err != nil {
		logger.Error("Failed to decode Vipps response", zap.Error(err), zap.String("id", id))
		return entities.Transaction{}, fmt.Errorf("failed to decode Vipps response: %w", err)
	}

	transaction := entities.Transaction{
		ID:              fmt.Sprintf("vipps_internal_%s", vt.TransactionID),
		ExternalID:      vt.TransactionID,
		Source:          consts.PAYMENT_SOURCE_VIPPS,
		Amount:          float64(vt.Amount) / 100, // Convert from øre to NOK
		Currency:        vt.Currency,
		Status:          statushelpers.NormalizeTransactionStatus(vt.Status, consts.PAYMENT_SOURCE_VIPPS),
		CreatedAt:       vt.TimeStamp,
		TransactionType: "mobile_payment",
		Description:     vt.TransactionText,
		PaymentMethod:   "vipps",
		Metadata:        map[string]string{"provider": "vipps", "order_id": vt.OrderID},
		CachedAt:        time.Now(),
	}

	return transaction, nil
}

// Helper function to get minimum of two integers
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// parseVippsResponse attempts to parse different Vipps API response formats
func (v *VippsClient) parseVippsResponse(bodyBytes []byte, endpoint string) ([]entities.Transaction, error) {
	var transactions []entities.Transaction

	// Try parsing as standard transaction response first
	var vippsResp VippsTransactionResponse
	if err := json.Unmarshal(bodyBytes, &vippsResp); err == nil && len(vippsResp.Transactions) > 0 {
		return v.convertVippsTransactions(vippsResp.Transactions), nil
	}

	// Try parsing as single transaction
	var singleTransaction VippsTransaction
	if err := json.Unmarshal(bodyBytes, &singleTransaction); err == nil && singleTransaction.TransactionID != "" {
		return v.convertVippsTransactions([]VippsTransaction{singleTransaction}), nil
	}

	// Try parsing Reports API responses (for transaction history)
	if strings.Contains(endpoint, "report") {
		if strings.Contains(endpoint, "transactions") {
			// Transaction report format
			var reportResp struct {
				Data []struct {
					TransactionID   string    `json:"transactionId"`
					OrderID         string    `json:"orderId"`
					Amount          int       `json:"amount"`
					Currency        string    `json:"currency"`
					TransactionTime time.Time `json:"transactionTime"`
					Status          string    `json:"status"`
					Description     string    `json:"description"`
				} `json:"data"`
			}

			if err := json.Unmarshal(bodyBytes, &reportResp); err == nil && len(reportResp.Data) > 0 {
				for _, tx := range reportResp.Data {
					transaction := entities.Transaction{
						ID:              fmt.Sprintf("vipps_report_%s", tx.TransactionID),
						ExternalID:      tx.TransactionID,
						Source:          consts.PAYMENT_SOURCE_VIPPS,
						Amount:          float64(tx.Amount) / 100,
						Currency:        tx.Currency,
						Status:          statushelpers.NormalizeTransactionStatus(tx.Status, consts.PAYMENT_SOURCE_VIPPS),
						CreatedAt:       tx.TransactionTime,
						TransactionType: "report_transaction",
						Description:     tx.Description,
						PaymentMethod:   "vipps",
						Metadata:        map[string]string{"provider": "vipps", "order_id": tx.OrderID, "source": "reports_api"},
						CachedAt:        time.Now(),
					}
					transactions = append(transactions, transaction)
				}
				if len(transactions) > 0 {
					return transactions, nil
				}
			}
		} else if strings.Contains(endpoint, "settlements") {
			// Settlement report format
			var settlementResp struct {
				Settlements []struct {
					SettlementID string `json:"settlementId"`
					Amount       int    `json:"amount"`
					Currency     string `json:"currency"`
					Date         string `json:"date"`
					Transactions []struct {
						TransactionID string `json:"transactionId"`
						Amount        int    `json:"amount"`
						OrderID       string `json:"orderId"`
					} `json:"transactions"`
				} `json:"settlements"`
			}

			if err := json.Unmarshal(bodyBytes, &settlementResp); err == nil && len(settlementResp.Settlements) > 0 {
				for _, settlement := range settlementResp.Settlements {
					settlementDate, _ := time.Parse("2006-01-02", settlement.Date)
					for _, tx := range settlement.Transactions {
						transaction := entities.Transaction{
							ID:              fmt.Sprintf("vipps_settlement_%s_%s", settlement.SettlementID, tx.TransactionID),
							ExternalID:      tx.TransactionID,
							Source:          consts.PAYMENT_SOURCE_VIPPS,
							Amount:          float64(tx.Amount) / 100,
							Currency:        settlement.Currency,
							Status:          statushelpers.NormalizeTransactionStatus("SETTLED", consts.PAYMENT_SOURCE_VIPPS),
							CreatedAt:       settlementDate,
							TransactionType: "settlement_transaction",
							Description:     fmt.Sprintf("Settlement %s", settlement.SettlementID),
							PaymentMethod:   "vipps",
							Metadata:        map[string]string{"provider": "vipps", "order_id": tx.OrderID, "settlement_id": settlement.SettlementID, "source": "reports_api"},
							CachedAt:        time.Now(),
						}
						transactions = append(transactions, transaction)
					}
				}
				if len(transactions) > 0 {
					return transactions, nil
				}
			}
		}
	}

	// Try parsing recurring agreements (different structure)
	if strings.Contains(endpoint, "recurring") {
		var recurringResp struct {
			Agreements []struct {
				ID      string `json:"id"`
				Status  string `json:"status"`
				Start   string `json:"start"`
				Charges []struct {
					ID     string `json:"id"`
					Amount int    `json:"amount"`
					Status string `json:"status"`
					Due    string `json:"due"`
				} `json:"charges"`
			} `json:"agreements"`
		}

		if err := json.Unmarshal(bodyBytes, &recurringResp); err == nil {
			for _, agreement := range recurringResp.Agreements {
				for _, charge := range agreement.Charges {
					if charge.Status == "CHARGED" || charge.Status == "COMPLETED" {
						dueTime, _ := time.Parse(time.RFC3339, charge.Due)
						transaction := entities.Transaction{
							ID:              fmt.Sprintf("vipps_recurring_%s_%s", agreement.ID, charge.ID),
							ExternalID:      charge.ID,
							Source:          consts.PAYMENT_SOURCE_VIPPS,
							Amount:          float64(charge.Amount) / 100,
							Currency:        "NOK",
							Status:          statushelpers.NormalizeTransactionStatus(charge.Status, consts.PAYMENT_SOURCE_VIPPS),
							CreatedAt:       dueTime,
							TransactionType: "recurring_payment",
							Description:     fmt.Sprintf("Recurring payment for agreement %s", agreement.ID),
							PaymentMethod:   "vipps",
							Metadata:        map[string]string{"provider": "vipps", "agreement_id": agreement.ID, "charge_id": charge.ID},
							CachedAt:        time.Now(),
						}
						transactions = append(transactions, transaction)
					}
				}
			}
			if len(transactions) > 0 {
				return transactions, nil
			}
		}
	}

	// Try parsing checkout sessions
	if strings.Contains(endpoint, "checkout") {
		var checkoutResp struct {
			Sessions []struct {
				SessionID string `json:"sessionId"`
				Amount    int    `json:"amount"`
				Status    string `json:"status"`
				Created   string `json:"created"`
			} `json:"sessions"`
		}

		if err := json.Unmarshal(bodyBytes, &checkoutResp); err == nil {
			for _, session := range checkoutResp.Sessions {
				if session.Status == "COMPLETED" || session.Status == "APPROVED" {
					createdTime, _ := time.Parse(time.RFC3339, session.Created)
					transaction := entities.Transaction{
						ID:              fmt.Sprintf("vipps_checkout_%s", session.SessionID),
						ExternalID:      session.SessionID,
						Source:          consts.PAYMENT_SOURCE_VIPPS,
						Amount:          float64(session.Amount) / 100,
						Currency:        "NOK",
						Status:          session.Status,
						CreatedAt:       createdTime,
						TransactionType: "checkout_payment",
						Description:     fmt.Sprintf("Checkout session %s", session.SessionID),
						PaymentMethod:   "vipps",
						Metadata:        map[string]string{"provider": "vipps", "session_id": session.SessionID},
						CachedAt:        time.Now(),
					}
					transactions = append(transactions, transaction)
				}
			}
			if len(transactions) > 0 {
				return transactions, nil
			}
		}
	}

	// If we can't parse it, return an error with the response for debugging
	return nil, fmt.Errorf("unable to parse Vipps response from endpoint %s. Response: %s",
		endpoint, string(bodyBytes[:minInt(500, len(bodyBytes))]))
}

// convertVippsTransactions converts VippsTransaction slice to entities.Transaction slice
func (v *VippsClient) convertVippsTransactions(vippsTransactions []VippsTransaction) []entities.Transaction {
	var transactions []entities.Transaction

	for _, vt := range vippsTransactions {
		transaction := entities.Transaction{
			ID:              fmt.Sprintf("vipps_internal_%s", vt.TransactionID),
			ExternalID:      vt.TransactionID,
			Source:          consts.PAYMENT_SOURCE_VIPPS,
			Amount:          float64(vt.Amount) / 100, // Convert from øre to NOK
			Currency:        vt.Currency,
			Status:          vt.Status,
			CreatedAt:       vt.TimeStamp,
			TransactionType: "mobile_payment",
			Description:     vt.TransactionText,
			PaymentMethod:   "vipps",
			Metadata:        map[string]string{"provider": "vipps", "order_id": vt.OrderID},
			CachedAt:        time.Now(),
		}
		transactions = append(transactions, transaction)
	}

	return transactions
}
