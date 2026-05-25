package calculator

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPMTCalculator_CalculatePayment(t *testing.T) {
	calc := NewPMTLoanCalculator()

	tests := []struct {
		name               string
		loanAmount         float64
		annualInterestRate float64
		numPayments        int32
		expectedPayment    float64
	}{
		{
			name:       "Standard 30-year mortgage at 5.5%",
			loanAmount: 100000.00, annualInterestRate: 5.5, numPayments: 360,
			expectedPayment: 567.79,
		},
		{
			name:       "15-year mortgage at 4.5%",
			loanAmount: 200000.00, annualInterestRate: 4.5, numPayments: 180,
			expectedPayment: 1529.99,
		},
		{
			name:       "5-year car loan at 6.0%",
			loanAmount: 25000.00, annualInterestRate: 6.0, numPayments: 60,
			expectedPayment: 483.32,
		},
		{
			name:       "Zero interest rate",
			loanAmount: 12000.00, annualInterestRate: 0.0, numPayments: 12,
			expectedPayment: 1000.00,
		},
		{
			name:       "High interest rate",
			loanAmount: 10000.00, annualInterestRate: 18.0, numPayments: 36,
			expectedPayment: 361.52,
		},
		{
			name:       "Single payment",
			loanAmount: 5000.00, annualInterestRate: 5.0, numPayments: 1,
			expectedPayment: 5020.83,
		},
		{
			name:       "Large loan amount",
			loanAmount: 1000000.00, annualInterestRate: 7.5, numPayments: 360,
			expectedPayment: 6992.15,
		},
		{
			name:       "Precision test",
			loanAmount: 123456.78, annualInterestRate: 4.375, numPayments: 360,
			expectedPayment: 616.40,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payment := calc.Calculate(tt.loanAmount, tt.annualInterestRate, tt.numPayments)
			diff := math.Abs(payment - tt.expectedPayment)
			assert.LessOrEqual(t, diff, 0.01,
				"Payment %.2f differs from expected %.2f by %.4f", payment, tt.expectedPayment, diff)
		})
	}
}
