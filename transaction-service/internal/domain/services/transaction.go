package services

import (
	"context"
	"log"
	"transaction-service/internal/domain/entities"
	"transaction-service/internal/domain/repositories"
	"transaction-service/kafka"
	"transaction-service/proto/gen"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	Deposit  entities.TransactionType = 0
	Withdraw entities.TransactionType = 1
)

type TransactionService interface {
	Deposit(ctx context.Context, req *gen.TransactionRequest) (*gen.TransactionResponse, error)
}

type TransactionServiceImpl struct {
	gen.UnimplementedTransactionServiceServer
	transactionRepo *repositories.PostgresTransactionRepository
	producer        *kafka.Producer
}

func NewTransactionService(transactionRepo *repositories.PostgresTransactionRepository, transactionProducer *kafka.Producer) *TransactionServiceImpl {
	return &TransactionServiceImpl{
		transactionRepo: transactionRepo,
		producer:        transactionProducer,
	}
}

func (s *TransactionServiceImpl) Deposit(ctx context.Context, req *gen.TransactionRequest) (*gen.TransactionResponse, error) {
	deposit := entities.Transaction{
		WalletID:       req.GetWalletId(),
		Amount:         req.GetAmount(),
		IdempotencyKey: req.GetIdempotencyKey(),
		Type:           Deposit,
	}

	if req.Amount <= 0 {
		return nil, status.Errorf(codes.FailedPrecondition, "amount must be positive")
	}

	// Start a transaction
	tx, err := s.transactionRepo.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback() // Roll back if not committed

	// Check idempotency within transaction
	existingID, err := s.transactionRepo.GetTxByKey(tx, req.GetIdempotencyKey())
	if err != nil {
		return nil, err
	}
	if existingID != "" {
		tx.Commit()
		return &gen.TransactionResponse{TransactionId: existingID}, nil
	}

	// Insert PENDING transaction
	txID, err := s.transactionRepo.InsertOne(tx, deposit)
	if err != nil {
		return nil, err
	}

	// Commit transaction before Kafka
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	// Publish to Kafka (non-transactional)
	err = s.producer.PublishDepositInitiated(ctx, deposit.WalletID, deposit.Amount, txID)
	if err != nil {
		log.Printf("Failed to publish to Kafka: %v; transaction %s is PENDING", err, txID)
		// TODO: Mark transaction as failed in Postgres
		return nil, err
	}

	return &gen.TransactionResponse{TransactionId: txID}, nil
}
