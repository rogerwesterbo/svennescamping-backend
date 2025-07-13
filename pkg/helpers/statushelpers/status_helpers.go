package statushelpers

import (
	"strings"

	"github.com/rogerwesterbo/svennescamping-backend/pkg/consts"
)

// NormalizeTransactionStatus converts provider-specific status to unified status
func NormalizeTransactionStatus(providerStatus, source string) string {
	// Normalize the input status (trim whitespace, convert to uppercase for comparison)
	normalizedStatus := strings.ToUpper(strings.TrimSpace(providerStatus))

	switch strings.ToLower(source) {
	case consts.PAYMENT_SOURCE_STRIPE:
		return normalizeStripeStatus(normalizedStatus)
	case consts.PAYMENT_SOURCE_VIPPS:
		return normalizeVippsStatus(normalizedStatus)
	case consts.PAYMENT_SOURCE_ZETTLE:
		return normalizeZettleStatus(normalizedStatus)
	default:
		// If unknown source, try to infer from common status names
		return inferStatusFromCommonNames(normalizedStatus)
	}
}

func normalizeStripeStatus(status string) string {
	// Check both uppercase and lowercase versions
	lowerStatus := strings.ToLower(status)

	if mappedStatus, exists := consts.StripeStatusMapping[lowerStatus]; exists {
		return mappedStatus
	}

	// Fallback for uppercase or mixed case
	for stripeStatus, unifiedStatus := range consts.StripeStatusMapping {
		if strings.EqualFold(status, stripeStatus) {
			return unifiedStatus
		}
	}

	return consts.TRANSACTION_STATUS_UNKNOWN
}

func normalizeVippsStatus(status string) string {
	if mappedStatus, exists := consts.VippsStatusMapping[status]; exists {
		return mappedStatus
	}

	// Try lowercase version
	lowerStatus := strings.ToLower(status)
	for vippsStatus, unifiedStatus := range consts.VippsStatusMapping {
		if strings.EqualFold(lowerStatus, vippsStatus) {
			return unifiedStatus
		}
	}

	return consts.TRANSACTION_STATUS_UNKNOWN
}

func normalizeZettleStatus(status string) string {
	if mappedStatus, exists := consts.ZettleStatusMapping[status]; exists {
		return mappedStatus
	}

	// Try lowercase version
	lowerStatus := strings.ToLower(status)
	for zettleStatus, unifiedStatus := range consts.ZettleStatusMapping {
		if strings.EqualFold(lowerStatus, zettleStatus) {
			return unifiedStatus
		}
	}

	return consts.TRANSACTION_STATUS_UNKNOWN
}

// inferStatusFromCommonNames tries to guess status from common payment terms
func inferStatusFromCommonNames(status string) string {
	lowerStatus := strings.ToLower(status)

	// Common success indicators
	if strings.Contains(lowerStatus, "success") ||
		strings.Contains(lowerStatus, "complete") ||
		strings.Contains(lowerStatus, "paid") ||
		strings.Contains(lowerStatus, "capture") {
		return consts.TRANSACTION_STATUS_SUCCEEDED
	}

	// Common pending indicators
	if strings.Contains(lowerStatus, "pending") ||
		strings.Contains(lowerStatus, "waiting") ||
		strings.Contains(lowerStatus, "initiated") {
		return consts.TRANSACTION_STATUS_PENDING
	}

	// Common processing indicators
	if strings.Contains(lowerStatus, "processing") ||
		strings.Contains(lowerStatus, "authorized") {
		return consts.TRANSACTION_STATUS_PROCESSING
	}

	// Common failure indicators
	if strings.Contains(lowerStatus, "fail") ||
		strings.Contains(lowerStatus, "reject") ||
		strings.Contains(lowerStatus, "decline") {
		return consts.TRANSACTION_STATUS_FAILED
	}

	// Common cancellation indicators
	if strings.Contains(lowerStatus, "cancel") ||
		strings.Contains(lowerStatus, "void") ||
		strings.Contains(lowerStatus, "abandon") {
		return consts.TRANSACTION_STATUS_CANCELLED
	}

	// Common refund indicators
	if strings.Contains(lowerStatus, "refund") {
		return consts.TRANSACTION_STATUS_REFUNDED
	}

	// Common expiration indicators
	if strings.Contains(lowerStatus, "expir") ||
		strings.Contains(lowerStatus, "timeout") {
		return consts.TRANSACTION_STATUS_EXPIRED
	}

	return consts.TRANSACTION_STATUS_UNKNOWN
}

// IsSuccessfulStatus checks if a unified status represents a successful payment
func IsSuccessfulStatus(status string) bool {
	return status == consts.TRANSACTION_STATUS_SUCCEEDED
}

// IsFinalStatus checks if a status represents a final state (no further changes expected)
func IsFinalStatus(status string) bool {
	finalStatuses := []string{
		consts.TRANSACTION_STATUS_SUCCEEDED,
		consts.TRANSACTION_STATUS_FAILED,
		consts.TRANSACTION_STATUS_CANCELLED,
		consts.TRANSACTION_STATUS_REFUNDED,
		consts.TRANSACTION_STATUS_EXPIRED,
	}

	for _, finalStatus := range finalStatuses {
		if status == finalStatus {
			return true
		}
	}

	return false
}

// GetStatusDescription returns a human-readable description of the status
func GetStatusDescription(status string) string {
	descriptions := map[string]string{
		consts.TRANSACTION_STATUS_PENDING:    "Payment is waiting to be processed",
		consts.TRANSACTION_STATUS_SUCCEEDED:  "Payment completed successfully",
		consts.TRANSACTION_STATUS_FAILED:     "Payment failed or was declined",
		consts.TRANSACTION_STATUS_CANCELLED:  "Payment was cancelled",
		consts.TRANSACTION_STATUS_REFUNDED:   "Payment was refunded",
		consts.TRANSACTION_STATUS_PROCESSING: "Payment is being processed",
		consts.TRANSACTION_STATUS_EXPIRED:    "Payment session expired",
		consts.TRANSACTION_STATUS_UNKNOWN:    "Payment status is unknown",
	}

	if description, exists := descriptions[status]; exists {
		return description
	}

	return "Unknown payment status"
}
