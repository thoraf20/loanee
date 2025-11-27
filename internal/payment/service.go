package payment

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/thoraf20/loanee/internal/loan"
	"github.com/thoraf20/loanee/internal/models"
)

type Service struct {
	repo        Repository
	loanService *loan.Service
	logger      zerolog.Logger
}

func NewService(repo Repository, loanService *loan.Service, logger zerolog.Logger) *Service {
	return &Service{
		repo:        repo,
		loanService: loanService,
		logger:      logger.With().Str("component", "payment_service").Logger(),
	}
}

type RepaymentRequest struct {
	Amount    float64 `json:"amount" binding:"required,gt=0"`
	Currency  string  `json:"currency" binding:"required,oneof=USD NGN"`
	Method    string  `json:"method"`
	Reference string  `json:"reference"`
}

type RepaymentResult struct {
	Loan      *models.Loan    `json:"loan"`
	Payment   *models.Payment `json:"payment"`
	Remaining float64         `json:"remaining_principal"`
}

func (s *Service) RecordRepayment(ctx context.Context, userID, loanID uuid.UUID, req RepaymentRequest) (*RepaymentResult, error) {
	if req.Amount <= 0 {
		return nil, fmt.Errorf("amount must be greater than zero")
	}

	loanSnapshot, breakdown, err := s.loanService.ApplyRepayment(ctx, loanID, userID, req.Amount)
	if err != nil {
		return nil, err
	}

	payment := &models.Payment{
		LoanID:          loanID,
		UserID:          userID,
		Amount:          req.Amount,
		Currency:        req.Currency,
		PrincipalAmount: breakdown.Principal,
		InterestAmount:  breakdown.Interest,
		PenaltyAmount:   breakdown.Penalty,
		Method:          req.Method,
		Reference:       req.Reference,
		Status:          models.PaymentCompleted,
		PaidAt:          time.Now(),
	}

	if err := s.repo.Create(ctx, payment); err != nil {
		return nil, err
	}

	return &RepaymentResult{
		Loan:      loanSnapshot,
		Payment:   payment,
		Remaining: loanSnapshot.PrincipalOutstanding,
	}, nil
}

func (s *Service) ListRepayments(ctx context.Context, loanID, userID uuid.UUID) ([]models.Payment, error) {
	loan, err := s.loanService.GetByID(ctx, loanID)
	if err != nil {
		return nil, err
	}
	if loan == nil || loan.UserID != userID {
		return nil, fmt.Errorf("loan not found")
	}
	return s.repo.ListByLoan(ctx, loanID)
}
