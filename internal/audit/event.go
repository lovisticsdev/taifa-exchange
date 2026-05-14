package audit

import (
	"time"

	"taifa-exchange/internal/platform/ids"
)

const SourceTaifaExchange = "taifa-exchange"

const (
	ResultSuccess = "SUCCESS"
	ResultFailure = "FAILURE"
	ResultAllowed = "ALLOWED"
	ResultDenied  = "DENIED"
)

const (
	ResourceExchangePolicy        = "EXCHANGE_POLICY"
	ResourceAuthorizationDecision = "AUTHORIZATION_DECISION"
	ResourceActorContext          = "ACTOR_CONTEXT"
	ResourceTaifaID               = "TAIFA_ID"
)

const (
	ActionCreatePolicy        = "CREATE_POLICY"
	ActionAuthorizeRequest    = "AUTHORIZE_REQUEST"
	ActionResolveActorContext = "RESOLVE_ACTOR_CONTEXT"
	ActionEvaluatePolicy      = "EVALUATE_POLICY"
	ActionPersistDecision     = "PERSIST_DECISION"
)

const (
	EventAuthorizationAllowed      = "exchange.authorization.allowed"
	EventAuthorizationDenied       = "exchange.authorization.denied"
	EventActorContextResolveFailed = "exchange.actor_context.resolve_failed"
	EventPolicyNotFound            = "exchange.policy.not_found"
	EventPolicyMatched             = "exchange.policy.matched"
	EventPolicySeeded              = "exchange.policy.seeded"
)

type Event struct {
	ID            string
	EventType     string
	SourceSystem  string
	SubjectID     string
	ActorID       string
	ResourceType  string
	ResourceID    string
	Action        string
	Result        string
	CorrelationID string
	Payload       map[string]any
	CreatedAt     time.Time
}

func (e Event) WithDefaults() Event {
	if e.ID == "" {
		e.ID = ids.NewAuditEventID()
	}

	if e.SourceSystem == "" {
		e.SourceSystem = SourceTaifaExchange
	}

	if e.CreatedAt.IsZero() {
		e.CreatedAt = time.Now().UTC()
	}

	if e.Payload == nil {
		e.Payload = map[string]any{}
	}

	return e
}
