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

	// Step 1: Check if the idempotency key already exists
	existingID, err := s.transactionRepo.GetByKey(req.GetIdempotencyKey())
	if err != nil {
		return nil, err
	}
	if existingID != "" {
		log.Printf("existing transaction with ID: %v", existingID)
		return &pb.TransactionResponse{TransactionId: existingID}, nil
	}

	// Step 2: Insert PENDING transaction
	txID, err := s.transactionRepo.InsertOne(deposit)
	if err != nil {
		return nil, err
	}

	//TODO: Add Race condition- idempotency key handling...

	// Step 3: Publish to Kafka
	err = s.producer.PublishDepositInitiated(ctx, deposit.WalletID, deposit.Amount, txID)
	if err != nil {
		log.Println("Failed to publish to Kafka:", err)
		return nil, err
	}

	return &pb.TransactionResponse{TransactionId: txID}, nil
}
