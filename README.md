# Loan Engine

A gRPC service that calculates monthly loan repayments using the PMT formula, built with Go and PostgreSQL 18.

## Architecture

```
client
  │
  ▼ gRPC (proto/loan/v1)
LoanCalculatorServer
  ├── PMTCalculator       — calculates monthly payment
  └── LoanRepository      — persists every request + response to Postgres
```

**Key design decisions:**

- **Decimal precision** — uses `shopspring/decimal` for PMT calculation to avoid float64 rounding errors on large loan amounts
- **Strategy pattern** — `CalculationMethod` enum in the proto allows new repayment formulas (flat-rate, reducing-balance, etc.) to be added without changing the handler
- **Structured logging** — `zap` for JSON logs in production, human-readable in development

## Prerequisites

- Go 1.21+
- PostgreSQL 16+
- [goose](https://github.com/pressly/goose) for migrations
- [grpcurl](https://github.com/fullstorydev/grpcurl) for testing

## Local Development

**1. Start the database**

```bash
docker-compose up -d
```

**2. Run migrations**

```bash
make db-up
```

**3. Start the server**

```bash
make run
```

The server starts on `localhost:50051`.

## Test the API

```bash
grpcurl -plaintext -d '{
  "loan_amount": 100000,
  "annual_interest_rate": 5.5,
  "num_payments": 360
}' localhost:50051 loan.v1.LoanCalculator/CalculatePayment
```

Response:

```json
{
  "monthlyPayment": 567.79,
  "calculationId": "2668ffc0-a222-48ba-99ce-5dd6b2a2a828"
}
```

### Calculation methods

The `calculation_method` field is optional and defaults to `PMT` (equal monthly payment / annuity).

| Value | Description |
|-------|-------------|
| `CALCULATION_METHOD_UNSPECIFIED` | defaults to PMT |
| `CALCULATION_METHOD_PMT` | equal monthly payment (standard annuity formula) |

## Run Tests

```bash
go test ./...
```

## Database Schema

```sql
CREATE TABLE calculation_records (
    id              UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    loan_amount     DECIMAL(20,2) NOT NULL,
    interest_rate   DECIMAL(8,4)  NOT NULL,
    num_payments    INTEGER       NOT NULL,
    monthly_payment DECIMAL(20,2) NOT NULL,
    created_at      TIMESTAMPTZ   NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

## Kubernetes Deployment

See [k8s/README.md](k8s/README.md) for full deployment instructions.
