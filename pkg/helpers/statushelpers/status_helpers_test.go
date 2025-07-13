package statushelpers

import (
	"testing"

	"github.com/rogerwesterbo/svennescamping-backend/pkg/consts"
)

func TestNormalizeTransactionStatus(t *testing.T) {
	tests := []struct {
		name           string
		providerStatus string
		source         string
		expected       string
	}{
		// Stripe tests
		{
			name:           "Stripe succeeded",
			providerStatus: "succeeded",
			source:         consts.PAYMENT_SOURCE_STRIPE,
			expected:       consts.TRANSACTION_STATUS_SUCCEEDED,
		},
		{
			name:           "Stripe requires_payment_method",
			providerStatus: "requires_payment_method",
			source:         consts.PAYMENT_SOURCE_STRIPE,
			expected:       consts.TRANSACTION_STATUS_PENDING,
		},
		{
			name:           "Stripe processing",
			providerStatus: "processing",
			source:         consts.PAYMENT_SOURCE_STRIPE,
			expected:       consts.TRANSACTION_STATUS_PROCESSING,
		},
		{
			name:           "Stripe canceled",
			providerStatus: "canceled",
			source:         consts.PAYMENT_SOURCE_STRIPE,
			expected:       consts.TRANSACTION_STATUS_CANCELLED,
		},

		// Vipps tests
		{
			name:           "Vipps CAPTURE",
			providerStatus: "CAPTURE",
			source:         consts.PAYMENT_SOURCE_VIPPS,
			expected:       consts.TRANSACTION_STATUS_SUCCEEDED,
		},
		{
			name:           "Vipps INITIATE",
			providerStatus: "INITIATE",
			source:         consts.PAYMENT_SOURCE_VIPPS,
			expected:       consts.TRANSACTION_STATUS_PENDING,
		},
		{
			name:           "Vipps RESERVE",
			providerStatus: "RESERVE",
			source:         consts.PAYMENT_SOURCE_VIPPS,
			expected:       consts.TRANSACTION_STATUS_PROCESSING,
		},
		{
			name:           "Vipps FAILED",
			providerStatus: "FAILED",
			source:         consts.PAYMENT_SOURCE_VIPPS,
			expected:       consts.TRANSACTION_STATUS_FAILED,
		},

		// Zettle tests
		{
			name:           "Zettle COMPLETED",
			providerStatus: "COMPLETED",
			source:         consts.PAYMENT_SOURCE_ZETTLE,
			expected:       consts.TRANSACTION_STATUS_SUCCEEDED,
		},
		{
			name:           "Zettle PENDING",
			providerStatus: "PENDING",
			source:         consts.PAYMENT_SOURCE_ZETTLE,
			expected:       consts.TRANSACTION_STATUS_PENDING,
		},
		{
			name:           "Zettle FAILED",
			providerStatus: "FAILED",
			source:         consts.PAYMENT_SOURCE_ZETTLE,
			expected:       consts.TRANSACTION_STATUS_FAILED,
		},

		// Case insensitive tests
		{
			name:           "Vipps lowercase capture",
			providerStatus: "capture",
			source:         consts.PAYMENT_SOURCE_VIPPS,
			expected:       consts.TRANSACTION_STATUS_SUCCEEDED,
		},
		{
			name:           "Stripe uppercase SUCCEEDED",
			providerStatus: "SUCCEEDED",
			source:         consts.PAYMENT_SOURCE_STRIPE,
			expected:       consts.TRANSACTION_STATUS_SUCCEEDED,
		},

		// Unknown provider inference
		{
			name:           "Unknown provider with success keyword",
			providerStatus: "payment_successful",
			source:         "unknown_provider",
			expected:       consts.TRANSACTION_STATUS_SUCCEEDED,
		},
		{
			name:           "Unknown provider with pending keyword",
			providerStatus: "pending_approval",
			source:         "unknown_provider",
			expected:       consts.TRANSACTION_STATUS_PENDING,
		},
		{
			name:           "Unknown provider with failure keyword",
			providerStatus: "payment_failed",
			source:         "unknown_provider",
			expected:       consts.TRANSACTION_STATUS_FAILED,
		},

		// Unknown status
		{
			name:           "Unknown status",
			providerStatus: "unknown_weird_status",
			source:         consts.PAYMENT_SOURCE_STRIPE,
			expected:       consts.TRANSACTION_STATUS_UNKNOWN,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(sub *testing.T) {
			result := NormalizeTransactionStatus(test.providerStatus, test.source)
			if result != test.expected {
				sub.Errorf("NormalizeTransactionStatus(%q, %q) = %q, want %q",
					test.providerStatus, test.source, result, test.expected)
			}
		})
	}
}

func TestIsSuccessfulStatus(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		expected bool
	}{
		{
			name:     "Succeeded status",
			status:   consts.TRANSACTION_STATUS_SUCCEEDED,
			expected: true,
		},
		{
			name:     "Failed status",
			status:   consts.TRANSACTION_STATUS_FAILED,
			expected: false,
		},
		{
			name:     "Pending status",
			status:   consts.TRANSACTION_STATUS_PENDING,
			expected: false,
		},
		{
			name:     "Cancelled status",
			status:   consts.TRANSACTION_STATUS_CANCELLED,
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(sub *testing.T) {
			result := IsSuccessfulStatus(test.status)
			if result != test.expected {
				sub.Errorf("IsSuccessfulStatus(%q) = %t, want %t", test.status, result, test.expected)
			}
		})
	}
}

func TestIsFinalStatus(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		expected bool
	}{
		{
			name:     "Succeeded (final)",
			status:   consts.TRANSACTION_STATUS_SUCCEEDED,
			expected: true,
		},
		{
			name:     "Failed (final)",
			status:   consts.TRANSACTION_STATUS_FAILED,
			expected: true,
		},
		{
			name:     "Cancelled (final)",
			status:   consts.TRANSACTION_STATUS_CANCELLED,
			expected: true,
		},
		{
			name:     "Refunded (final)",
			status:   consts.TRANSACTION_STATUS_REFUNDED,
			expected: true,
		},
		{
			name:     "Expired (final)",
			status:   consts.TRANSACTION_STATUS_EXPIRED,
			expected: true,
		},
		{
			name:     "Pending (not final)",
			status:   consts.TRANSACTION_STATUS_PENDING,
			expected: false,
		},
		{
			name:     "Processing (not final)",
			status:   consts.TRANSACTION_STATUS_PROCESSING,
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(sub *testing.T) {
			result := IsFinalStatus(test.status)
			if result != test.expected {
				sub.Errorf("IsFinalStatus(%q) = %t, want %t", test.status, result, test.expected)
			}
		})
	}
}

func TestGetStatusDescription(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		expected string
	}{
		{
			name:     "Succeeded description",
			status:   consts.TRANSACTION_STATUS_SUCCEEDED,
			expected: "Payment completed successfully",
		},
		{
			name:     "Failed description",
			status:   consts.TRANSACTION_STATUS_FAILED,
			expected: "Payment failed or was declined",
		},
		{
			name:     "Unknown status description",
			status:   "weird_status",
			expected: "Unknown payment status",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(sub *testing.T) {
			result := GetStatusDescription(test.status)
			if result != test.expected {
				sub.Errorf("GetStatusDescription(%q) = %q, want %q", test.status, result, test.expected)
			}
		})
	}
}
