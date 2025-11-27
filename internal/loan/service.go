package loan

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/thoraf20/loanee/config"
	"github.com/thoraf20/loanee/internal/models"
)

type Service struct {
	repo   Repository
	cfg    *config.Config
	logger zerolog.Logger
}

func NewService(repo Repository, cfg *config.Config, logger zerolog.Logger) *Service {
	return &Service{
		repo:   repo,
		cfg:    cfg,
		logger: logger.With().Str("component", "loan_service").Logger(),
	}
}

func (s *Service) CreateFromCollateral(ctx context.Context, collateral *models.Collateral) (*models.Loan, error) {
	if collateral == nil {
		return nil, fmt.Errorf("collateral is nil")
	}

	loan := &models.Loan{
		UserID:               collateral.UserID,
		CollateralID:         collateral.ID,
		AmountRequested:      collateral.FiatAmount,
		AmountApproved:       collateral.FiatAmount,
		PrincipalOutstanding: 0,
		InterestRate:         s.cfg.Loan.DefaultInterestRate,
		DurationMonths:       12,
		Status:               "pending",
	}

	if err := s.repo.Create(ctx, loan); err != nil {
		return nil, err
	}
	return loan, nil
}

func (s *Service) ListUserLoans(ctx context.Context, userID uuid.UUID) ([]models.Loan, error) {
	return s.repo.ListByUser(ctx, userID)
}

func (s *Service) ListAllLoans(ctx context.Context) ([]models.Loan, error) {
	return s.repo.ListAll(ctx)
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*models.Loan, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) ApproveLoan(ctx context.Context, id uuid.UUID, amount float64) (*models.Loan, error) {
	loan, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if loan == nil {
		return nil, fmt.Errorf("loan not found")
	}
	if amount <= 0 {
		amount = loan.AmountRequested
	}
	loan.AmountApproved = amount
	loan.Status = "approved"
	if err := s.repo.Update(ctx, loan); err != nil {
		return nil, err
	}
	return loan, nil
}

func (s *Service) DisburseLoan(ctx context.Context, id uuid.UUID) (*models.Loan, error) {
	loan, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if loan == nil {
		return nil, fmt.Errorf("loan not found")
	}
	now := time.Now()
	loan.DisbursedAt = &now
	loan.PrincipalOutstanding = loan.AmountApproved
	loan.Status = "active"
	nextDue := now.Add(time.Duration(s.cfg.Loan.RepaymentFrequencyDays) * 24 * time.Hour)
	loan.NextDueDate = &nextDue
	if err := s.repo.Update(ctx, loan); err != nil {
		return nil, err
	}
	return loan, nil
}

type RepaymentBreakdown struct {
	Principal float64 `json:"principal"`
	Interest  float64 `json:"interest"`
	Penalty   float64 `json:"penalty"`
}

func (s *Service) ApplyRepayment(ctx context.Context, loanID, userID uuid.UUID, amount float64) (*models.Loan, *RepaymentBreakdown, error) {
	loan, err := s.repo.GetByID(ctx, loanID)
	if err != nil {
		return nil, nil, err
	}
	if loan == nil {
		return nil, nil, fmt.Errorf("loan not found")
	}
	if loan.UserID != userID {
		return nil, nil, fmt.Errorf("loan does not belong to user")
	}
	if loan.PrincipalOutstanding <= 0 {
		return loan, &RepaymentBreakdown{}, nil
	}

	interestDue := s.currentInterestDue(loan)
	penaltyDue := s.currentPenaltyDue(loan)
	breakdown := &RepaymentBreakdown{}
	remaining := amount

	if penaltyDue > 0 && remaining > 0 {
		pay := math.Min(penaltyDue, remaining)
		breakdown.Penalty = pay
		remaining -= pay
		penaltyDue -= pay
	}

	if interestDue > 0 && remaining > 0 {
		pay := math.Min(interestDue, remaining)
		breakdown.Interest = pay
		remaining -= pay
		interestDue -= pay
	}

	if remaining > 0 {
		principalPay := math.Min(loan.PrincipalOutstanding, remaining)
		breakdown.Principal = principalPay
		remaining -= principalPay
		loan.PrincipalOutstanding -= principalPay
	}

	now := time.Now()
	loan.PenaltyAccrued = penaltyDue
	loan.TotalRepaid += amount - remaining
	loan.LastPaymentAt = &now

	if loan.PrincipalOutstanding <= 0.01 {
		loan.PrincipalOutstanding = 0
		loan.Status = "repaid"
		loan.NextDueDate = nil
	} else {
		nextDue := now.Add(time.Duration(s.cfg.Loan.RepaymentFrequencyDays) * 24 * time.Hour)
		loan.NextDueDate = &nextDue
		if loan.PenaltyAccrued > 0 {
			loan.Status = "delinquent"
		} else {
			loan.Status = "active"
		}
	}

	if err := s.repo.Update(ctx, loan); err != nil {
		return nil, nil, err
	}

	return loan, breakdown, nil
}

func (s *Service) currentInterestDue(loan *models.Loan) float64 {
	monthlyRate := (loan.InterestRate / 100) / 12
	return loan.PrincipalOutstanding * monthlyRate
}

func (s *Service) currentPenaltyDue(loan *models.Loan) float64 {
	if loan.NextDueDate == nil {
		return loan.PenaltyAccrued
	}
	grace := time.Duration(s.cfg.Loan.GracePeriodDays) * 24 * time.Hour
	if time.Since(loan.NextDueDate.Add(grace)) <= 0 {
		return loan.PenaltyAccrued
	}
	daysLate := int(time.Since(loan.NextDueDate.Add(grace)).Hours() / 24)
	if daysLate < 0 {
		daysLate = 0
	}
	dailyRate := (s.cfg.Loan.PenaltyAPR / 100) / 365
	penalty := loan.PrincipalOutstanding * dailyRate * float64(daysLate)
	return loan.PenaltyAccrued + penalty
}
