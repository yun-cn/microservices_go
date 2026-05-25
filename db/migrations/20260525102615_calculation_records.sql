-- +goose Up
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS calculation_records (
                                                   id              UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    loan_amount     DECIMAL(20,2) NOT NULL,
    interest_rate   DECIMAL(8,4)  NOT NULL,
    num_payments    INTEGER       NOT NULL,
    monthly_payment DECIMAL(20,2) NOT NULL,
    created_at      TIMESTAMPTZ   NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT loan_amount_positive     CHECK (loan_amount > 0),
    CONSTRAINT interest_rate_valid      CHECK (interest_rate >= 0 AND interest_rate <= 100),
    CONSTRAINT num_payments_positive    CHECK (num_payments > 0),
    CONSTRAINT monthly_payment_positive CHECK (monthly_payment > 0)
);

CREATE INDEX idx_calculation_records_created_at ON calculation_records(created_at DESC);

-- +goose Down
DROP INDEX IF EXISTS idx_calculation_records_created_at;
DROP TABLE IF EXISTS calculation_records;
DROP EXTENSION IF EXISTS "pgcrypto";