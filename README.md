# Crypto-backed Loan API

A Go REST API that allows users to deposit digital assets (ETH, BNB, etc.) as collateral and receive fiat loans calculated from a Loan-To-Value (LTV) ratio. Repayments are scheduled; late repayments incur penalties. After full repayment, users can request release of their collateral.

## Features
- User registration & JWT authentication (basic scaffold)
- Collateral recording (crypto deposits)
- Loan creation with LTV checks and repayment schedule generation
- Repayment endpoints (partial & full)
- Penalty calculation for late payments (configurable)
- Collateral release request & admin approval flow
- PostgreSQL persistence via GORM
- DB migrations via `golang-migrate`
- Swagger (swaggo) docs support

## Tech stack
- Go (>=1.20)
- PostgreSQL
- GORM
- Zerolog (logging)
- Viper (config)
- golang-migrate (migrations)
- swaggo/swag (API docs)
- golangci-lint, air, testify for dev/test tooling


## Getting started (dev)

### Prerequisites
- Go >= 1.20 installed
- PostgreSQL running (local or remote)
- `migrate` CLI for migrations (optional but recommended)
- `air` for hot reload (optional)
- `swag` for swagger generation (optional)

### Env / Config
Create `.env` (or set env vars) with the following values: