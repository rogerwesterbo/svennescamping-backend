package consts

// Environment and general config
var (
	DEVELOPMENT     = "DEVELOPMENT"
	CORS_ORIGINS    = "CORS_ORIGINS"
	USER_EMAILS     = "USER_EMAILS"
	ADMIN_EMAILS    = "ADMIN_EMAILS"
	PRICES_CSV_PATH = "PRICES_CSV_PATH"
)

// Stripe configuration
var (
	STRIPE_APIKEY     = "STRIPE_APIKEY"
	STRIPE_WEBHOOKKEY = "STRIPE_WEBHOOKKEY"
	STRIPE_WEBHOOKURL = "STRIPE_WEBHOOKURL"
	STRIPE_APIURL     = "STRIPE_APIURL"
	STRIPE_APIVERSION = "STRIPE_APIVERSION"
)

// Vipps configuration
var (
	VIPPS_SUBSCRIPTION_KEY       = "VIPPS_SUBSCRIPTION_KEY"
	VIPPS_APIURL                 = "VIPPS_APIURL"
	VIPPS_CLIENT_ID              = "VIPPS_CLIENT_ID"
	VIPPS_SECRET                 = "VIPPS_SECRET"
	VIPPS_MERCHANT_SERIAL_NUMBER = "VIPPS_MERCHANT_SERIAL_NUMBER"
)

// Zettle configuration
var (
	ZETTLE_APIKEY    = "ZETTLE_APIKEY"
	ZETTLE_APIURL    = "ZETTLE_APIURL"
	ZETTLE_CLIENT_ID = "ZETTLE_CLIENT_ID"
	ZETTLE_SECRET    = "ZETTLE_SECRET"
)

// Payment sources
var (
	PAYMENT_SOURCE_STRIPE = "stripe"
	PAYMENT_SOURCE_VIPPS  = "vipps"
	PAYMENT_SOURCE_ZETTLE = "zettle"
)

// Transaction status constants - unified across all payment providers
var (
	TRANSACTION_STATUS_PENDING    = "pending"    // Payment initiated but not completed
	TRANSACTION_STATUS_SUCCEEDED  = "succeeded"  // Payment completed successfully
	TRANSACTION_STATUS_FAILED     = "failed"     // Payment failed or was declined
	TRANSACTION_STATUS_CANCELLED  = "cancelled"  // Payment was cancelled by user or system
	TRANSACTION_STATUS_REFUNDED   = "refunded"   // Payment was refunded (partial or full)
	TRANSACTION_STATUS_PROCESSING = "processing" // Payment is being processed
	TRANSACTION_STATUS_EXPIRED    = "expired"    // Payment session expired without completion
	TRANSACTION_STATUS_UNKNOWN    = "unknown"    // Status could not be determined
)

// Stripe status mapping to unified status
var StripeStatusMapping = map[string]string{
	"requires_payment_method": TRANSACTION_STATUS_PENDING,
	"requires_confirmation":   TRANSACTION_STATUS_PENDING,
	"requires_action":         TRANSACTION_STATUS_PENDING,
	"processing":              TRANSACTION_STATUS_PROCESSING,
	"requires_capture":        TRANSACTION_STATUS_PROCESSING,
	"succeeded":               TRANSACTION_STATUS_SUCCEEDED,
	"canceled":                TRANSACTION_STATUS_CANCELLED,
	"payment_failed":          TRANSACTION_STATUS_FAILED,
	"refunded":                TRANSACTION_STATUS_REFUNDED,
	"partially_refunded":      TRANSACTION_STATUS_REFUNDED,
}

// Vipps status mapping to unified status
var VippsStatusMapping = map[string]string{
	"INITIATE":  TRANSACTION_STATUS_PENDING,
	"REGISTER":  TRANSACTION_STATUS_PENDING,
	"RESERVE":   TRANSACTION_STATUS_PROCESSING,
	"CAPTURE":   TRANSACTION_STATUS_SUCCEEDED,
	"SALE":      TRANSACTION_STATUS_SUCCEEDED,
	"CANCEL":    TRANSACTION_STATUS_CANCELLED,
	"VOID":      TRANSACTION_STATUS_CANCELLED,
	"REFUND":    TRANSACTION_STATUS_REFUNDED,
	"FAILED":    TRANSACTION_STATUS_FAILED,
	"REJECTED":  TRANSACTION_STATUS_FAILED,
	"EXPIRED":   TRANSACTION_STATUS_EXPIRED,
	"ABANDONED": TRANSACTION_STATUS_CANCELLED,
}

// Zettle status mapping to unified status
var ZettleStatusMapping = map[string]string{
	"PENDING":    TRANSACTION_STATUS_PENDING,
	"COMPLETED":  TRANSACTION_STATUS_SUCCEEDED,
	"FAILED":     TRANSACTION_STATUS_FAILED,
	"CANCELLED":  TRANSACTION_STATUS_CANCELLED,
	"REFUNDED":   TRANSACTION_STATUS_REFUNDED,
	"VOIDED":     TRANSACTION_STATUS_CANCELLED,
	"PROCESSING": TRANSACTION_STATUS_PROCESSING,
	"AUTHORIZED": TRANSACTION_STATUS_PROCESSING,
	"CAPTURED":   TRANSACTION_STATUS_SUCCEEDED,
}

// Transaction limits
var (
	TRANSACTION_LIMIT_MIN     = 1
	TRANSACTION_LIMIT_DEFAULT = 25
	TRANSACTION_LIMIT_MAX     = 1000
)
