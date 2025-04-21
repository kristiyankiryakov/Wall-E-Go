package cmd

import (
	"capture/internal/config"
	"capture/internal/db"
	"capture/internal/transaction"
	"context"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
	"time"
)

func NewCaptureCmd() *cobra.Command {
	cmdInstance := &cobra.Command{
		Use:   "capture",
		Short: "Capture simulates a capture process via fake SFTP serve",
		Long:  `Capture simulates a capture process via fake SFTP serve`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ExecuteCaptureAuthTransactions()

			return nil
		},
	}
	cmdInstance.Flags().String("config", "", "Path to the config file (eg. config.yaml)")
	_ = viper.BindPFlag("config", cmdInstance.Flags().Lookup("config"))
	return cmdInstance
}

func ExecuteCaptureAuthTransactions() {
	ctx := context.Background()
	cfg := config.NewConfig()

	if err := db.InitDB(cfg.DatabaseUrl); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.DB.Close()

	repository := transaction.NewPostgresTransactionRepository(db.DB)

	start := time.Now().Add(-2 * time.Hour).UTC().Format("2006-01-02 15:04:05")
	log.Printf("Start time: %s", start)
	end := time.Now().Add(-1 * time.Hour).UTC().Format("2006-01-02 15:04:05")
	log.Printf("Start time: %s", end)

	transactions, err := repository.GetTransactionsForProcessing(ctx, start, end)
	if err != nil {
		log.Fatalf("Failed to get authorized transactions: %v", err)
	}

	log.Printf("Found %d authorized transactions", len(transactions))
}
