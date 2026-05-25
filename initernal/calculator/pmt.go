package calculator

import (
	"github.com/shopspring/decimal"
)

type PMTCalculator struct{}

// NewPMTLoanCalculator NewPMTCalculator creates a new PMT calculator
func NewPMTLoanCalculator() *PMTCalculator { return &PMTCalculator{} }

// Calculate CalculatePayment calculates the monthly payment using the PMT formula.
//
// PMT Formula: PMT = (PV * r * (1 + r)^n) / ((1 + r)^n - 1)
// Where:
//   - PV = Present Value (loan amount)
//   - r = Periodic interest rate (annual rate / 12 / 100)
//   - n = Number of payments
//
// Special case: When interest rate is 0, PMT = PV / n
//
// Returns the monthly payment amount with 2 decimal precision.
func (c *PMTCalculator) Calculate(loanAmount, annualInterestRate float64, numPayments int32) float64 {

	// Convert inputs to decimal for precision
	pv := decimal.NewFromFloat(loanAmount)
	rate := decimal.NewFromFloat(annualInterestRate)
	n := decimal.NewFromInt32(numPayments)

	// Zero interest: PMT = PV / n
	if rate.IsZero() {
		result, _ := pv.Div(n).Round(2).Float64()
		return result
	}

	// PMT = PV * r * (1+r)^n / ((1+r)^n - 1)
	r := rate.Div(decimal.NewFromInt(12)).Div(decimal.NewFromInt(100))
	pow := decimal.NewFromInt(1).Add(r).Pow(n)
	result, _ := pv.Mul(r).Mul(pow).Div(pow.Sub(decimal.NewFromInt(1))).Round(2).Float64()
	return result
}
