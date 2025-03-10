package service

import (
	"context"
	"log"
	"transaction-service/internal/data"
	"transaction-service/kafka"
	pb "transaction-service/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TransactionService interface {
	Deposit(ctx context.Context, req *pb.TransactionRequest) (*pb.TransactionResponse, error)
}

type TransactionServiceImpl struct {
	pb.UnimplementedTransactionServiceServer
	transactionRepo data.TransactionRepository
	producer        *kafka.Producer
}

func NewTransactionService(transactionRepo data.TransactionRepository, transactionProducer *kafka.Producer) *TransactionServiceImpl {
	return &TransactionServiceImpl{
		transactionRepo: transactionRepo,
		producer:        transactionProducer,
	}
}

func (s *TransactionServiceImpl) Deposit(ctx context.Context, req *pb.TransactionRequest) (*pb.TransactionResponse, error) {
	deposit := data.TransactionRequest{
		WalletID:       req.GetWalletId(),
		Amount:         req.GetAmount(),
		IdempotencyKey: req.GetIdempotencyKey(),
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
		return &pb.TransactionResponse{TransactionId: existingID}, nil
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

	return &pb.TransactionResponse{TransactionId: txID}, nil
}
