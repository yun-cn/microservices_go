package calculator

// LoanCalculator defines the interface for all loan calculation types.
// Each implementation represents a different repayment structure.
type LoanCalculator interface {
	Calculate(loanAmount, annualInterestRate float64, numPayments int32) float64
}
