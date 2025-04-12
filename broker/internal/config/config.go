package config

// Config holds application configuration
type Config struct {
	AuthHost        string
	WalletHost      string
	TransactionHost string
	ServerPort      string
	JwtKey          string
}

// NewConfig creates a new configuration
func NewConfig() *Config {
	serverPort := "8080"
	authHost := "localhost:50051"
	walletHost := "localhost:50052"
	transactionHost := "localhost:50053"
	jwtKey := "MIGbMBAGByqGSM49AgEGBSuBBAAjA4GGAAQAwczAzq0G4TRw8Qj69FRxnuV380IQ"

	return &Config{
		authHost,
		walletHost,
		transactionHost,
		serverPort,
		jwtKey,
	}
}
