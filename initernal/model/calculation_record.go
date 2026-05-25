package model

import "time"

type CalculationRecord struct {
	ID             string    `json:"id" gorm:"primaryKey"`
	LoanAmount     float64   `json:"loan_amount" gorm:"type:decimal(20,2);not null"`
	InterestRate   float64   `json:"interest_rate" gorm:"type:decimal(8,4);not null"`
	NumPayments    int32     `json:"num_payments" gorm:"not null"`
	MonthlyPayment float64   `json:"monthly_payment" gorm:"type:decimal(20,2);not null"`
	CreatedAt      time.Time `json:"created_at" gorm:"autoCreateTime"`
}

func (CalculationRecord) TableName() string {
	return "calculation_records"
}
