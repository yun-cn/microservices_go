# Time Log

Total estimated time: 6h

| Area | Estimated Hours | Notes |
|------|----------------|-------|
| Proto / gRPC API | 1.5h           | Designing the service contract, message types, and CalculationMethod enum for extensibility |
| PMT Calculation | 0.5h           | Researching precision trade-offs; chose `shopspring/decimal` over `float64` to avoid rounding errors on large loan amounts |
| Database (schema + repository) | 0.5h           | Schema design, GORM setup, goose migrations — straightforward |
| Tests | 1h             | Setting up mock repository, table-driven tests for validation and strategy pattern |
| Kubernetes | 2h             | Limited prior experience — debugged image loading, postgres CrashLoopBackOff from PVC version mismatch, gRPC health probe setup |
| Tooling & Config (Makefile, Docker, env) | 0.5h           | Makefile targets, Dockerfile multi-stage build, local dev environment setup |
| Documentation | 0.5h           | README, k8s deployment guide, this log |

## What Took Longer Than Expected

**Proto / gRPC API (~1.5h):** More thought went into the API design than expected — specifically around the `CalculationMethod` enum and how to structure request/response messages to be extensible without breaking existing clients.


**Kubernetes (~1.5h):** Most of the extra time was spent on:
- Understanding how local images work with Docker Desktop's k8s (no registry push needed)
- Diagnosing a PostgreSQL `CrashLoopBackOff` caused by a data format mismatch between postgres versions on a reused PVC
- Configuring the gRPC health check protocol (`grpc.health.v1`) so readiness/liveness probes report correctly

## What Took Shorter Than Expected

**Database (~0.5h):** Schema design was clear from the requirements. GORM and goose are well-documented and the single-table setup required minimal complexity.

**Tooling & Config (~0.5h):** More time than expected was spent getting everything running end-to-end — debugging the Dockerfile multi-stage build (binary path conflict), fixing golangci-lint installation across platforms, and wiring up the Makefile targets to work reliably for local dev and k8s deployment.

**PMT Calculation (~0.5h):** Once the precision approach was decided, the implementation was straightforward. The formula is well-documented and `shopspring/decimal` handled the edge cases cleanly.
