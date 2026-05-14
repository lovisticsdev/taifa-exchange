
# TaifaExchange

Exchange and request-boundary service for Taifa Republic, a synthetic digital public infrastructure range.

TaifaExchange owns route policies, authorization decisions, actor-context validation through TaifaID, cross-system request correlation, and exchange audit events.

## Current status

Batch 2 provides PostgreSQL readiness, schema migrations, and audit outbox support.

Implemented:

```text
GET /healthz
GET /readyz
correlation IDs
request logging
panic recovery
graceful shutdown
PostgreSQL connection pool
database readiness
exchange schema migration
audit outbox insert helper
```

vvvv
