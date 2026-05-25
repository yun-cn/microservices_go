# Kubernetes Deployment

Tested on Docker Desktop with Kubernetes enabled. Should work on any local cluster (minikube, kind, etc.).

## Prerequisites

- Docker Desktop with Kubernetes enabled
- `kubectl` configured to the local cluster
- `grpcurl` for testing

## First-time Setup

### 1. Build the image

```bash
docker build -t loan-engine:latest .
```

### 2. Apply config and secrets

```bash
kubectl apply -f k8s/config.yaml
```

### 3. Start PostgreSQL

```bash
kubectl apply -f k8s/postgres.yaml
```

Wait until it is ready:

```bash
kubectl get pods -l app=postgres -w
# Wait for STATUS = Running, READY = 1/1, then Ctrl+C
```

### 4. Run migrations and start the app

```bash
kubectl apply -f k8s/app.yaml
```

Check migration completed:

```bash
kubectl logs -l app=loan-engine,component=migration
# Should see: "goose: successfully migrated database to version: ..."
```

Check app is running:

```bash
kubectl get pods -l app=loan-engine,component!=migration
# Should see: STATUS = Running, READY = 1/1
```

## Test the API

Forward the gRPC port to your local machine:

```bash
kubectl port-forward service/loan-engine-service 50051:50051
```

In another terminal:

```bash
grpcurl -plaintext -d '{
  "loan_amount": 100000,
  "annual_interest_rate": 5.5,
  "num_payments": 360
}' localhost:50051 loan.v1.LoanCalculator/CalculatePayment
```

## Redeploying After Code Changes

```bash
docker build -t loan-engine:latest .
kubectl rollout restart deployment/loan-engine
kubectl rollout status deployment/loan-engine
```

## Useful Commands

```bash
# View all resources
kubectl get pods,svc,job

# View app logs
kubectl logs -l app=loan-engine --tail=50

# View database logs
kubectl logs -l app=postgres --tail=50

# Connect to the database directly
kubectl exec -it $(kubectl get pod -l app=postgres -o name) -- \
  psql -U loan_engine_app -d loan_engine
```

## Full Reset

```bash
kubectl delete -f k8s/app.yaml
kubectl delete -f k8s/postgres.yaml
kubectl delete -f k8s/config.yaml
kubectl delete pvc postgres-pvc
```
