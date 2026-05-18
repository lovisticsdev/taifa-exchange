
# TaifaExchange

Exchange and request-boundary service for Taifa Republic, a synthetic digital public infrastructure range.

TaifaExchange owns route policies, authorization decisions, actor-context validation through TaifaID, cross-system request correlation, and exchange audit events.

## Foundation status

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
exchange schema migrations
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
integration smoke tests
Docker image
Docker Compose local workflow
```

## Boundary

TaifaExchange does not own identity data.

TaifaID remains the source of truth for:

```text
persons
organizations
memberships
roles
credentials
organization capabilities
actor-context resolution
identity audit events
```

TaifaExchange owns:

```text
route policies
authorization decisions
request-boundary audit events
cross-system request correlation evidence
```

## Service flow

```text
request
→ Exchange extracts Bearer token
→ Exchange calls TaifaID actor-context resolution
→ Exchange calls TaifaID organization-capability lookup
→ Exchange evaluates route policy
→ Exchange persists authorization decision
→ Exchange writes audit outbox event
→ Exchange returns ALLOW or DENY
```

## Environment

Local defaults:

```powershell
$env:TAIFA_EXCHANGE_HTTP_ADDR=":8081"

$env:TAIFA_EXCHANGE_DATABASE_DSN="host=localhost port=5432 user=taifa password=taifa_dev_password dbname=taifa_exchange sslmode=disable"
$env:TAIFA_EXCHANGE_DATABASE_CONNECT_TIMEOUT="15s"

$env:TAIFA_EXCHANGE_TAIFA_ID_BASE_URL="http://localhost:8080"
$env:TAIFA_EXCHANGE_TAIFA_ID_TIMEOUT="10s"

$env:TAIFA_EXCHANGE_SEED_TIMEOUT="5m"
```

AWS RDS pattern:

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

Run:

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

Canonical policies:

```text
POL-CARE-CLAIM-CREATE-CLINICIAN
POL-CARE-CLAIM-SUBMIT-CLAIMS-OFFICER
POL-TAX-CONTRIBUTION-SUBMIT-EMPLOYER
POL-TAX-CONTRIBUTION-READ-TAX-OFFICER
POL-PAY-INSTRUCTION-CREATE-PAY-OPERATOR
POL-OBSERVE-SECURITY-EVENT-READ-ANALYST
```

## Run locally

Start TaifaID first on `:8080`.

Then start TaifaExchange:

```powershell
go run ./cmd/taifa-exchange
```

Default address:

```text
:8081
```

Health checks:

```powershell
Invoke-RestMethod http://localhost:8081/healthz
Invoke-RestMethod http://localhost:8081/readyz
```

Expected readiness after database and TaifaID are configured:

```text
database = ok
taifa_id = ok
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

Expected seeded result for `clinician.seed`:

```text
decision = ALLOW
matched_policy_id = POL-CARE-CLAIM-CREATE-CLINICIAN
```

Expected denial test for `clinician.seed` against payment instruction creation:

```text
status = 403
error.code = FORBIDDEN
database deny_reason = missing_required_role
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

Fail-closed denial reasons include:

```text
missing_token
missing_organization_id
missing_target_system
unsupported_target_system
missing_method
unsupported_method
missing_route
missing_operation
actor_context_resolve_failed
organization_capabilities_lookup_failed
policy_not_found
policy_disabled
explicit_policy_deny
missing_required_role
missing_required_capability
```

## Integration smoke tests

Preconditions:

```text
TaifaID running on :8080
TaifaExchange running on :8081
taifa_exchange database migrated
canonical exchange policies seeded
TaifaID canonical seed data available
```

Default smoke-test environment:

```powershell
$env:TAIFA_EXCHANGE_TEST_BASE_URL="http://localhost:8081"
$env:TAIFA_EXCHANGE_TEST_TAIFA_ID_BASE_URL="http://localhost:8080"
$env:TAIFA_EXCHANGE_TEST_USERNAME="clinician.seed"
$env:TAIFA_EXCHANGE_TEST_PASSWORD="ExampleDevPass123!"
$env:TAIFA_EXCHANGE_TEST_ORGANIZATION_ID="ORG-HP-CLINIC"
```

Run:

```powershell
go test -count=1 ./tests/integration
```

Force the test to fail instead of skip when services are unavailable:

```powershell
$env:TAIFA_EXCHANGE_RUN_INTEGRATION_TESTS="true"
go test -count=1 ./tests/integration
```

The smoke test verifies:

```text
clinician.seed can authorize TAIFA_CARE care.claim.create
clinician.seed cannot authorize TAIFA_PAY pay.instruction.create
```

## Docker build

```powershell
docker build -t taifa-exchange:local .
```

## Docker Compose local workflow

This starts a local PostgreSQL database, applies migrations, seeds canonical policies, and starts TaifaExchange.

TaifaID must already be running on the host at `http://localhost:8080`.

```powershell
docker compose up --build
```

Check readiness:

```powershell
Invoke-RestMethod http://localhost:8081/readyz
```

Expected:

```text
database = ok
taifa_id = ok
```

Stop:

```powershell
docker compose down
```

Remove local database volume:

```powershell
docker compose down -v
```

## Validate

```powershell
go fmt ./...
go mod tidy
go test ./...
docker build -t taifa-exchange:local .
```

## Repository checkpoint

```text
Batch 0: repo skeleton                              done
Batch 0.1: compile-safe placeholders                done
Batch 1: bootable HTTP service scaffold             done
Batch 2: PostgreSQL schema and audit outbox         done
Batch 3: TaifaID client                             done
Batch 4: policy model and evaluator                 done
Batch 5: authorization decision API                 done
Batch 6: canonical exchange policy seed command     done
Batch 7: integration smoke tests                    done
Batch 8: Docker and README stabilization            done
```
