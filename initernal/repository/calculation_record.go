package repository

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/yun-cn/lone-engine/initernal/model"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Repository interface {
	SaveCalculation(ctx context.Context, cal *model.CalculationRecord) (string, error)
}

type CalculationRecordRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

func NewCalculationRecordRepository(db *gorm.DB, logger *zap.Logger) *CalculationRecordRepository {
	return &CalculationRecordRepository{db: db, logger: logger}
}

// SaveCalculation saves a loan calculation to the database and returns the UUID
func (r *CalculationRecordRepository) SaveCalculation(ctx context.Context, calc *model.CalculationRecord) (string, error) {
	if calc.ID == "" {
		calc.ID = uuid.New().String()
	}

	if result := r.db.WithContext(ctx).Create(calc); result.Error != nil {
		return "", fmt.Errorf("failed to save calculation: %w", result.Error)
	}

	return calc.ID, nil
}
