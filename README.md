
# TaifaExchange

Exchange and request-boundary service for Taifa Republic, a synthetic digital public infrastructure range.

TaifaExchange owns route policies, authorization decisions, actor-context validation through TaifaID, cross-system request correlation, and exchange audit events.

## Current status

Batch 6 provides the canonical exchange policy seed command.

Implemented:

```text
GET /healthz
GET /readyz
POST /api/v1/exchange/authorize
correlation IDs
request logging
panic recovery
graceful shutdown
PostgreSQL connection pool
database readiness
exchange schema migration
audit outbox insert helper
TaifaID readiness client
TaifaID actor-context resolution client
TaifaID organization-capability client
policy model
policy repository
policy evaluator
policy service
policy evaluator unit tests
authorization decision persistence
authorization audit events
canonical exchange policy seeding
seed audit evidence
```

Not yet implemented:

```text
integration smoke tests
Docker runtime
```

## Authorization API

Endpoint:

```text
POST /api/v1/exchange/authorize
```

Header:

```text
Authorization: Bearer <jwt-from-taifa-id>
```

Request:

```json
{
  "organization_id": "ORG-HP-CLINIC",
  "target_system": "TAIFA_CARE",
  "route": "/api/v1/claims",
  "method": "POST",
  "operation": "care.claim.create"
}
```

After canonical policies are seeded, the clinician seed actor should receive `ALLOW` for `care.claim.create`.

## Canonical policies

The seed command inserts these policies idempotently:

```text
POL-CARE-CLAIM-CREATE-CLINICIAN
POL-CARE-CLAIM-SUBMIT-CLAIMS-OFFICER
POL-TAX-CONTRIBUTION-SUBMIT-EMPLOYER
POL-TAX-CONTRIBUTION-READ-TAX-OFFICER
POL-PAY-INSTRUCTION-CREATE-PAY-OPERATOR
POL-OBSERVE-SECURITY-EVENT-READ-ANALYST
```

## Policy semantics

```text
route matching: exact route_pattern match
method matching: exact HTTP method match
operation matching: exact operation match
target matching: exact target_system match
role requirement: actor must have at least one required role
capability requirement: organization must have all required capabilities
default decision: DENY
```

## Environment

```powershell
$env:TAIFA_EXCHANGE_HTTP_ADDR=":8081"

$env:TAIFA_EXCHANGE_DATABASE_DSN="host=localhost port=5432 user=taifa password=taifa_dev_password dbname=taifa_exchange sslmode=disable"
$env:TAIFA_EXCHANGE_DATABASE_CONNECT_TIMEOUT="15s"

$env:TAIFA_EXCHANGE_TAIFA_ID_BASE_URL="http://localhost:8080"
$env:TAIFA_EXCHANGE_TAIFA_ID_TIMEOUT="10s"

$env:TAIFA_EXCHANGE_SEED_TIMEOUT="5m"
```

For AWS RDS:

```powershell
$env:TAIFA_EXCHANGE_DB_HOST=$env:TAIFA_ID_DB_HOST
$env:TAIFA_EXCHANGE_DB_PASSWORD=$env:TAIFA_ID_DB_PASSWORD
$env:TAIFA_EXCHANGE_DATABASE_DSN="host=$env:TAIFA_EXCHANGE_DB_HOST port=5432 user=taifa password=$env:TAIFA_EXCHANGE_DB_PASSWORD dbname=taifa_exchange sslmode=require"
$env:TAIFA_EXCHANGE_DATABASE_CONNECT_TIMEOUT="15s"
$env:TAIFA_EXCHANGE_SEED_TIMEOUT="5m"
```

Do not commit real credentials.

## Migrations

Apply migrations to AWS RDS:

```powershell
Get-Content migrations\000001_init.up.sql | docker run -i --rm `
  -e PGPASSWORD="$env:TAIFA_EXCHANGE_DB_PASSWORD" `
  postgres:latest `
  psql "host=$env:TAIFA_EXCHANGE_DB_HOST port=5432 dbname=taifa_exchange user=taifa sslmode=require"

Get-Content migrations\000002_allow_unknown_decision_inputs.up.sql | docker run -i --rm `
  -e PGPASSWORD="$env:TAIFA_EXCHANGE_DB_PASSWORD" `
  postgres:latest `
  psql "host=$env:TAIFA_EXCHANGE_DB_HOST port=5432 dbname=taifa_exchange user=taifa sslmode=require"
```

Verify tables:

```powershell
docker run --rm `
  -e PGPASSWORD="$env:TAIFA_EXCHANGE_DB_PASSWORD" `
  postgres:latest `
  psql "host=$env:TAIFA_EXCHANGE_DB_HOST port=5432 dbname=taifa_exchange user=taifa sslmode=require" `
  -c "\dt"
```

Expected tables:

```text
audit_outbox
exchange_authorization_decisions
exchange_policies
exchange_policy_capabilities
exchange_policy_roles
```

## Seed

Run the seed command:

```powershell
go run ./cmd/taifa-exchange-seed
```

Expected first run:

```text
taifa-exchange seed completed policies=6 roles=6 capabilities=6 audit_events=6 seed_timeout=5m0s
```

Expected second run:

```text
taifa-exchange seed completed policies=0 roles=0 capabilities=0 audit_events=0 seed_timeout=5m0s
```

Verify seeded policies:

```powershell
docker run --rm `
  -e PGPASSWORD="$env:TAIFA_EXCHANGE_DB_PASSWORD" `
  postgres:latest `
  psql "host=$env:TAIFA_EXCHANGE_DB_HOST port=5432 dbname=taifa_exchange user=taifa sslmode=require" `
  -c "SELECT id, target_system, method, route_pattern, operation, effect, status FROM exchange_policies ORDER BY id;"
```

## Run

Start TaifaID first on `:8080`.

Then start TaifaExchange:

```powershell
go run ./cmd/taifa-exchange
```

Default address:

```text
:8081
```

## Health

```powershell
Invoke-RestMethod http://localhost:8081/healthz
Invoke-RestMethod http://localhost:8081/readyz
```

Expected readiness after database and TaifaID are configured:

```text
database = ok
taifa_id = ok
```

## Validate

```powershell
go fmt ./...
go mod tidy
go test ./...
```

## Boundary

TaifaExchange does not own identity data.

TaifaID remains the source of truth for persons, credentials, memberships, roles, and organization capabilities.

TaifaExchange owns route-boundary policy decisions and exchange audit evidence.
