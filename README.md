
# TaifaExchange

Exchange and request-boundary service for Taifa Republic, a synthetic digital public infrastructure range.

TaifaExchange owns route policies, authorization decisions, actor-context validation through TaifaID, cross-system request correlation, and exchange audit events.

## Current status

Batch 5 provides the authorization decision API.

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
```
