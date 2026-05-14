package exchange

import (
	"strings"
	"time"
)

const (
	DecisionAllow = "ALLOW"
	DecisionDeny  = "DENY"
)

const (
	TargetSystemUnknown      = "UNKNOWN"
	TargetSystemTaifaID      = "TAIFA_ID"
	TargetSystemTaifaCare    = "TAIFA_CARE"
	TargetSystemTaifaTax     = "TAIFA_TAX"
	TargetSystemTaifaPay     = "TAIFA_PAY"
	TargetSystemTaifaObserve = "TAIFA_OBSERVE"
	TargetSystemTaifaCitizen = "TAIFA_CITIZEN"
)

const (
	MethodUnknown = "UNKNOWN"
	MethodGet     = "GET"
	MethodPost    = "POST"
	MethodPut     = "PUT"
	MethodPatch   = "PATCH"
	MethodDelete  = "DELETE"
)

const (
	ErrorCodeValidation   = "VALIDATION_ERROR"
	ErrorCodeUnauthorized = "UNAUTHORIZED"
	ErrorCodeForbidden    = "FORBIDDEN"
	ErrorCodeUnavailable  = "UPSTREAM_UNAVAILABLE"
)

const (
	DenyReasonMissingToken                         = "missing_token"
	DenyReasonMissingOrganizationID                = "missing_organization_id"
	DenyReasonMissingTargetSystem                  = "missing_target_system"
	DenyReasonUnsupportedTargetSystem              = "unsupported_target_system"
	DenyReasonMissingMethod                        = "missing_method"
	DenyReasonUnsupportedMethod                    = "unsupported_method"
	DenyReasonMissingRoute                         = "missing_route"
	DenyReasonMissingOperation                     = "missing_operation"
	DenyReasonTaifaIDClientNotConfigured           = "taifa_id_client_not_configured"
	DenyReasonActorContextResolveFailed            = "actor_context_resolve_failed"
	DenyReasonOrganizationCapabilitiesLookupFailed = "organization_capabilities_lookup_failed"
	DenyReasonPolicyEvaluationFailed               = "policy_evaluation_failed"
)

type AuthorizeInput struct {
	Token         string
	CorrelationID string
	Request       AuthorizeRequest
}

type AuthorizeResult struct {
	DecisionRecord DecisionRecord
	Response       *AuthorizeResponse
	HTTPStatusCode int
	ErrorCode      string
	ErrorMessage   string
}

type DecisionRecord struct {
	ID                       string
	Decision                 string
	TargetSystem             string
	Route                    string
	Method                   string
	Operation                string
	OrganizationID           string
	ActorContextID           string
	PersonID                 string
	CredentialID             string
	SessionID                string
	MatchedPolicyID          string
	DenyReason               string
	RequestCorrelationID     string
	TaifaIDCorrelationID     string
	ActorRoles               []string
	OrganizationCapabilities []string
	RequiredRoles            []string
	RequiredCapabilities     []string
	CreatedAt                time.Time
}

func NormalizeTargetSystemForDecision(value string) string {
	value = strings.ToUpper(strings.TrimSpace(value))

	switch value {
	case TargetSystemTaifaID,
		TargetSystemTaifaCare,
		TargetSystemTaifaTax,
		TargetSystemTaifaPay,
		TargetSystemTaifaObserve,
		TargetSystemTaifaCitizen:
		return value
	default:
		return TargetSystemUnknown
	}
}

func IsKnownTargetSystem(value string) bool {
	return NormalizeTargetSystemForDecision(value) != TargetSystemUnknown
}

func NormalizeMethodForDecision(value string) string {
	value = strings.ToUpper(strings.TrimSpace(value))

	switch value {
	case MethodGet, MethodPost, MethodPut, MethodPatch, MethodDelete:
		return value
	default:
		return MethodUnknown
	}
}

func IsKnownMethod(value string) bool {
	return NormalizeMethodForDecision(value) != MethodUnknown
}

func NormalizeRoute(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "UNKNOWN"
	}

	return value
}

func NormalizeOperation(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "UNKNOWN"
	}

	return value
}

func NormalizeOrganizationID(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "UNKNOWN"
	}

	return value
}
