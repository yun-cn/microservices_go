package server

import (
	loanv1 "github.com/yun-cn/lone-engine/api/proto/loan/v1"
	"github.com/yun-cn/lone-engine/initernal/calculator"
	"github.com/yun-cn/lone-engine/initernal/logger"
	"github.com/yun-cn/lone-engine/initernal/repository"
)

type LoanCalculatorServer struct {
	loanv1.UnimplementedLoanCalculatorServer
	calculators map[loanv1.CalculationMethod]calculator.LoanCalculator
	repository  repository.Repository
	logger      *logger.Logger
}

func NewLoanCalculatorServer(repo repository.Repository, log *logger.Logger) *LoanCalculatorServer {
	pmt := calculator.NewPMTLoanCalculator()
	return &LoanCalculatorServer{
		calculators: map[loanv1.CalculationMethod]calculator.LoanCalculator{
			loanv1.CalculationMethod_CALCULATION_METHOD_UNSPECIFIED: pmt,
			loanv1.CalculationMethod_CALCULATION_METHOD_PMT:         pmt,
		},
		repository: repo,
		logger:     log,
	}
}
