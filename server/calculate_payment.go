package server

import (
	"context"
	loanv1 "github.com/yun-cn/lone-engine/api/proto/loan/v1"
	"github.com/yun-cn/lone-engine/initernal/model"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *LoanCalculatorServer) CalculatePayment(ctx context.Context, req *loanv1.CalculatePaymentRequest) (*loanv1.CalculatePaymentResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}
	if req.LoanAmount <= 0 {
		return nil, status.Error(codes.InvalidArgument, "loan amount must be greater than 0")
	}
	if req.LoanAmount > 100_000_000 {
		return nil, status.Errorf(codes.InvalidArgument, "loan amount %.2f exceeds maximum allowed", req.LoanAmount)
	}
	if req.AnnualInterestRate < 0 || req.AnnualInterestRate > 100 {
		return nil, status.Error(codes.InvalidArgument, "interest rate must be between 0 and 100")
	}
	if req.NumPayments <= 0 {
		return nil, status.Error(codes.InvalidArgument, "number of payments must be greater than 0")
	}
	if req.NumPayments > 600 {
		return nil, status.Error(codes.InvalidArgument, "number of payments cannot exceed 600")
	}
	s.logger.Info("Processing loan calculation",
		zap.Float64("loan_amount", req.LoanAmount),
		zap.Float64("interest_rate", req.AnnualInterestRate),
		zap.Int32("num_payments", req.NumPayments),
	)

	calc, ok := s.calculators[req.CalculationMethod]
	if !ok {
		return nil, status.Errorf(codes.InvalidArgument, "unsupported calculation method: %v", req.CalculationMethod)
	}

	monthlyPayment := calc.Calculate(req.LoanAmount, req.AnnualInterestRate, req.NumPayments)

	calcID, err := s.repository.SaveCalculation(ctx, &model.CalculationRecord{
		LoanAmount:     req.LoanAmount,
		InterestRate:   req.AnnualInterestRate,
		NumPayments:    req.NumPayments,
		MonthlyPayment: monthlyPayment,
	})
	if err != nil {
		s.logger.Error("Failed to save calculation", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to save calculation: %v", err)
	}

	s.logger.Info("Calculation completed",
		zap.String("calculation_id", calcID),
		zap.Float64("monthly_payment", monthlyPayment),
	)

	return &loanv1.CalculatePaymentResponse{
		MonthlyPayment: monthlyPayment,
		CalculationId:  calcID,
	}, nil
}
