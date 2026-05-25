package server

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	loanv1 "github.com/yun-cn/lone-engine/api/proto/loan/v1"
	"github.com/yun-cn/lone-engine/initernal/logger"
	"github.com/yun-cn/lone-engine/initernal/model"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) GetCalculateByID(ctx context.Context, id string) (*model.CalculationRecord, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.CalculationRecord), args.Error(1)
}

func (m *MockRepository) GetAll(ctx context.Context, page, pageSize int) ([]*model.CalculationRecord, int64, error) {
	args := m.Called(ctx, page, pageSize)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*model.CalculationRecord), int64(args.Int(1)), args.Error(2)
}

func (m *MockRepository) SaveCalculation(ctx context.Context, calc *model.CalculationRecord) (string, error) {
	args := m.Called(ctx, calc)
	return args.Get(0).(string), args.Error(1)
}

func getTestLogger() *logger.Logger {
	log, _ := logger.NewLogger("test")
	return log
}

func newTestServer(repo *MockRepository) *LoanCalculatorServer {
	return NewLoanCalculatorServer(repo, getTestLogger())
}

func TestCalculatePayment_InvalidInput(t *testing.T) {
	tests := []struct {
		name string
		req  *loanv1.CalculatePaymentRequest
		code codes.Code
	}{
		{
			name: "nil request",
			req:  nil,
			code: codes.InvalidArgument,
		},
		{
			name: "zero loan amount",
			req:  &loanv1.CalculatePaymentRequest{LoanAmount: 0, AnnualInterestRate: 5.5, NumPayments: 360},
			code: codes.InvalidArgument,
		},
		{
			name: "negative loan amount",
			req:  &loanv1.CalculatePaymentRequest{LoanAmount: -1000, AnnualInterestRate: 5.5, NumPayments: 360},
			code: codes.InvalidArgument,
		},
		{
			name: "loan amount exceeds maximum",
			req:  &loanv1.CalculatePaymentRequest{LoanAmount: 100_000_001, AnnualInterestRate: 5.5, NumPayments: 360},
			code: codes.InvalidArgument,
		},
		{
			name: "negative interest rate",
			req:  &loanv1.CalculatePaymentRequest{LoanAmount: 100000, AnnualInterestRate: -1, NumPayments: 360},
			code: codes.InvalidArgument,
		},
		{
			name: "interest rate exceeds 100",
			req:  &loanv1.CalculatePaymentRequest{LoanAmount: 100000, AnnualInterestRate: 101, NumPayments: 360},
			code: codes.InvalidArgument,
		},
		{
			name: "zero payments",
			req:  &loanv1.CalculatePaymentRequest{LoanAmount: 100000, AnnualInterestRate: 5.5, NumPayments: 0},
			code: codes.InvalidArgument,
		},
		{
			name: "negative payments",
			req:  &loanv1.CalculatePaymentRequest{LoanAmount: 100000, AnnualInterestRate: 5.5, NumPayments: -1},
			code: codes.InvalidArgument,
		},
		{
			name: "payments exceed maximum",
			req:  &loanv1.CalculatePaymentRequest{LoanAmount: 100000, AnnualInterestRate: 5.5, NumPayments: 601},
			code: codes.InvalidArgument,
		},
		{
			name: "unsupported calculation method",
			req:  &loanv1.CalculatePaymentRequest{LoanAmount: 100000, AnnualInterestRate: 5.5, NumPayments: 360, CalculationMethod: 99},
			code: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &MockRepository{}
			s := newTestServer(repo)

			_, err := s.CalculatePayment(context.Background(), tt.req)

			assert.Error(t, err)
			st, ok := status.FromError(err)
			assert.True(t, ok)
			assert.Equal(t, tt.code, st.Code())
			repo.AssertNotCalled(t, "SaveCalculation")
		})
	}
}

func TestCalculatePayment_Success(t *testing.T) {
	repo := &MockRepository{}
	repo.On("SaveCalculation", mock.Anything, mock.MatchedBy(func(calc *model.CalculationRecord) bool {
		return calc.LoanAmount == 100000 &&
			calc.InterestRate == 5.5 &&
			calc.NumPayments == 360 &&
			calc.MonthlyPayment > 0
	})).Return("test-uuid-1234", nil)

	s := newTestServer(repo)
	resp, err := s.CalculatePayment(context.Background(), &loanv1.CalculatePaymentRequest{
		LoanAmount:         100000,
		AnnualInterestRate: 5.5,
		NumPayments:        360,
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "test-uuid-1234", resp.CalculationId)
	assert.InDelta(t, 567.79, resp.MonthlyPayment, 0.01)
	repo.AssertExpectations(t)
}

func TestCalculatePayment_ExplicitPMTMethod(t *testing.T) {
	repo := &MockRepository{}
	repo.On("SaveCalculation", mock.Anything, mock.Anything).Return("test-uuid-pmt", nil)

	s := newTestServer(repo)
	resp, err := s.CalculatePayment(context.Background(), &loanv1.CalculatePaymentRequest{
		LoanAmount:         100000,
		AnnualInterestRate: 5.5,
		NumPayments:        360,
		CalculationMethod:  loanv1.CalculationMethod_CALCULATION_METHOD_PMT,
	})

	assert.NoError(t, err)
	assert.InDelta(t, 567.79, resp.MonthlyPayment, 0.01)
	repo.AssertExpectations(t)
}

func TestCalculatePayment_ZeroInterestRate(t *testing.T) {
	repo := &MockRepository{}
	repo.On("SaveCalculation", mock.Anything, mock.MatchedBy(func(calc *model.CalculationRecord) bool {
		return calc.InterestRate == 0 && calc.MonthlyPayment == 1000.00
	})).Return("test-uuid-5678", nil)

	s := newTestServer(repo)
	resp, err := s.CalculatePayment(context.Background(), &loanv1.CalculatePaymentRequest{
		LoanAmount:         12000,
		AnnualInterestRate: 0,
		NumPayments:        12,
	})

	assert.NoError(t, err)
	assert.InDelta(t, 1000.00, resp.MonthlyPayment, 0.01)
	repo.AssertExpectations(t)
}

func TestCalculatePayment_RepositoryError(t *testing.T) {
	repo := &MockRepository{}
	repo.On("SaveCalculation", mock.Anything, mock.Anything).
		Return("", errors.New("db connection lost"))

	s := newTestServer(repo)
	_, err := s.CalculatePayment(context.Background(), &loanv1.CalculatePaymentRequest{
		LoanAmount:         100000,
		AnnualInterestRate: 5.5,
		NumPayments:        360,
	})

	assert.Error(t, err)
	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.Internal, st.Code())
	repo.AssertExpectations(t)
}
