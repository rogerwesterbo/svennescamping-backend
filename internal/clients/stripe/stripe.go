package stripe

import (
	"context"
	"fmt"
	"time"

	"github.com/rogerwesterbo/svennescamping-backend/pkg/consts"
	"github.com/rogerwesterbo/svennescamping-backend/pkg/entities"
	"github.com/rogerwesterbo/svennescamping-backend/pkg/helpers/statushelpers"
	"github.com/rogerwesterbo/svennescamping-backend/pkg/interfaces"
	"github.com/rogerwesterbo/svennescamping-backend/pkg/logger"
	"github.com/stripe/stripe-go/v78"
	"github.com/stripe/stripe-go/v78/charge"
	"go.uber.org/zap"
)

type StripeClient struct {
	APIKey string
	ctx    context.Context
}

// Compile-time check to ensure StripeClient implements Transactions interface
var _ interfaces.Transactions = (*StripeClient)(nil)

func NewStripeClient(apiKey string) *StripeClient {
	stripe.Key = apiKey
	return &StripeClient{APIKey: apiKey}
}

func (s *StripeClient) GetLatestTransactions(ctx context.Context, limit int) ([]entities.Transaction, error) {
	params := &stripe.ChargeListParams{}
	params.Limit = stripe.Int64(int64(limit))
	params.Context = ctx

	i := charge.List(params)

	var transactions []entities.Transaction
	for i.Next() {
		ch := i.Charge()
		transaction := entities.Transaction{
			ID:              fmt.Sprintf("stripe_internal_%s", ch.ID),
			ExternalID:      ch.ID,
			Source:          consts.PAYMENT_SOURCE_STRIPE,
			Amount:          float64(ch.Amount) / 100, // Stripe beløp er i cent
			Currency:        string(ch.Currency),
			Status:          statushelpers.NormalizeTransactionStatus(string(ch.Status), consts.PAYMENT_SOURCE_STRIPE),
			CreatedAt:       time.Unix(ch.Created, 0),
			TransactionType: string(ch.PaymentMethodDetails.Type),
			Description:     ch.Description,
			ReceiptURL:      ch.ReceiptURL,
			Metadata:        ch.Metadata,
			//Data:            ch,
			TransferData: ch.TransferData,
			CachedAt:     time.Now(),
		}

		transactions = append(transactions, transaction)
	}

	if err := i.Err(); err != nil {
		logger.Error("Error retrieving charges", zap.Error(err))
		return nil, fmt.Errorf("error retrieving charges: %w", err)
	}

	return transactions, nil
}

func (s *StripeClient) GetTransactionByID(ctx context.Context, id string) (entities.Transaction, error) {
	params := &stripe.ChargeParams{}
	params.Context = ctx

	ch, err := charge.Get(id, params)
	if err != nil {
		logger.Error("Error retrieving charge by ID", zap.Error(err), zap.String("id", id))
		return entities.Transaction{}, fmt.Errorf("error retrieving charge by ID: %w", err)
	}

	transaction := entities.Transaction{
		ID:              fmt.Sprintf("stripe_internal_%s", ch.ID),
		ExternalID:      ch.ID,
		Source:          consts.PAYMENT_SOURCE_STRIPE,
		Amount:          float64(ch.Amount) / 100, // Stripe beløp er i cent
		Currency:        string(ch.Currency),
		Status:          statushelpers.NormalizeTransactionStatus(string(ch.Status), consts.PAYMENT_SOURCE_STRIPE),
		CreatedAt:       time.Unix(ch.Created, 0),
		TransactionType: string(ch.PaymentMethodDetails.Type),
		Description:     ch.Description,
		ReceiptURL:      ch.ReceiptURL,
		Metadata:        ch.Metadata,
		Data:            ch,
		TransferData:    ch.TransferData,
		CachedAt:        time.Now(),
	}

	return transaction, nil
}
