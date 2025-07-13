package clients

import (
	"context"
	"time"

	"github.com/rogerwesterbo/svennescamping-backend/internal/cache"
	"github.com/rogerwesterbo/svennescamping-backend/internal/clients/stripe"
	"github.com/rogerwesterbo/svennescamping-backend/internal/clients/vipps"
	"github.com/rogerwesterbo/svennescamping-backend/internal/clients/zettle"
	"github.com/rogerwesterbo/svennescamping-backend/internal/repository"
	"github.com/rogerwesterbo/svennescamping-backend/internal/services"
	"github.com/rogerwesterbo/svennescamping-backend/pkg/consts"
	"github.com/rogerwesterbo/svennescamping-backend/pkg/interfaces"
	"github.com/rogerwesterbo/svennescamping-backend/pkg/logger"
	"github.com/spf13/viper"
)

var (
	StripeClient          *stripe.StripeClient
	VippsClient           *vipps.VippsClient
	ZettleClient          *zettle.ZettleClient
	Cache                 interfaces.Cache
	TransactionRepository interfaces.TransactionRepository
)

func InitializeClients() {
	// Initialize cache with 24h default expiration and 1h cleanup interval
	Cache = cache.NewInMemoryCache(24*time.Hour, 1*time.Hour)

	// Initialize Stripe client
	stripeAPIKey := viper.GetString(consts.STRIPE_APIKEY)
	if stripeAPIKey != "" {
		StripeClient = stripe.NewStripeClient(stripeAPIKey)
	}

	// Initialize Vipps client
	vippsSubscriptionKey := viper.GetString(consts.VIPPS_SUBSCRIPTION_KEY)
	vippsAPIURL := viper.GetString(consts.VIPPS_APIURL)
	vippsClientID := viper.GetString(consts.VIPPS_CLIENT_ID)
	vippsSecret := viper.GetString(consts.VIPPS_SECRET)
	vippsMerchantSerialNumber := viper.GetString(consts.VIPPS_MERCHANT_SERIAL_NUMBER)
	if vippsSubscriptionKey != "" {
		VippsClient = vipps.NewVippsClient(vippsSubscriptionKey, vippsAPIURL, vippsClientID, vippsSecret, vippsMerchantSerialNumber)
	}

	// Initialize Zettle client
	zettleAPIKey := viper.GetString(consts.ZETTLE_APIKEY)
	zettleAPIURL := viper.GetString(consts.ZETTLE_APIURL)
	zettleClientID := viper.GetString(consts.ZETTLE_CLIENT_ID)
	zettleSecret := viper.GetString(consts.ZETTLE_SECRET)
	if zettleAPIKey != "" {
		ZettleClient = zettle.NewZettleClient(zettleAPIKey, zettleAPIURL, zettleClientID, zettleSecret)
	}

	// Initialize repository with all available clients
	TransactionRepository = repository.NewTransactionRepository(
		Cache,
		StripeClient,
		VippsClient,
		ZettleClient,
	)

	// Initialize transaction services through the services package
	services.InitializeTransactionServices(
		Cache,
		TransactionRepository,
		StripeClient,
		VippsClient,
		ZettleClient,
	)

	logger.Info("All clients and services initialized successfully")
}

// StartBackgroundFetching starts the background data fetching from all providers
func StartBackgroundFetching(ctx context.Context) {
	services.StartBackgroundFetching(ctx)
}

// StopBackgroundFetching stops the background data fetching
func StopBackgroundFetching() {
	services.StopBackgroundFetching()
}

// IsBackgroundFetchingRunning returns true if background fetching is active
func IsBackgroundFetchingRunning() bool {
	return services.IsBackgroundFetchingRunning()
}
