package settings

import (
	"os"
	"path/filepath"

	"github.com/rogerwesterbo/svennescamping-backend/pkg/consts"
	"github.com/rogerwesterbo/svennescamping-backend/pkg/logger"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var (
	Version = "0.0.0"
	Commit  = "localdev"
)

func Init() {
	// Load environment files in order: .env first, then .env.local (overrides .env)
	loadEnvFile(".env")
	loadEnvFile(".env.local")

	// Set default values
	viper.SetDefault(consts.DEVELOPMENT, false)
	viper.SetDefault(consts.STRIPE_APIKEY, "")
	viper.SetDefault(consts.STRIPE_WEBHOOKKEY, "")
	viper.SetDefault(consts.STRIPE_WEBHOOKURL, "https://example.com/webhook")
	viper.SetDefault(consts.STRIPE_APIURL, "https://api.stripe.com")
	viper.SetDefault(consts.STRIPE_APIVERSION, "2020-08-27")
	viper.SetDefault(consts.CORS_ORIGINS, "http://localhost:5173")

	// Load environment variables from the process (highest priority)
	viper.AutomaticEnv()
}

// loadEnvFile loads environment variables from a file if it exists
func loadEnvFile(filename string) {
	// Get the working directory
	wd, err := os.Getwd()
	if err != nil {
		logger.Warn("Failed to get working directory", zap.Error(err))
		return
	}

	// Try to find the env file in the current directory or project root
	envPath := filepath.Join(wd, filename)
	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		// If not found in current directory, try project root
		// Go up directories until we find go.mod (project root)
		dir := wd
		for {
			goModPath := filepath.Join(dir, "go.mod")
			if _, err := os.Stat(goModPath); err == nil {
				envPath = filepath.Join(dir, filename)
				break
			}

			parent := filepath.Dir(dir)
			if parent == dir {
				// Reached filesystem root
				logger.Debug("Environment file not found", zap.String("filename", filename))
				return
			}
			dir = parent
		}
	}

	// Check if the file exists
	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		logger.Debug("Environment file not found", zap.String("path", envPath))
		return
	}

	// Configure viper to read from the env file
	viper.SetConfigFile(envPath)
	viper.SetConfigType("env")

	if err := viper.ReadInConfig(); err != nil {
		logger.Warn("Failed to read environment file",
			zap.String("path", envPath),
			zap.Error(err))
		return
	}

	logger.Info("Loaded environment file", zap.String("path", envPath))
}
