package policy

import (
	"strings"
	"time"
)

const (
	TargetSystemTaifaID      = "TAIFA_ID"
	TargetSystemTaifaCare    = "TAIFA_CARE"
	TargetSystemTaifaTax     = "TAIFA_TAX"
	TargetSystemTaifaPay     = "TAIFA_PAY"
	TargetSystemTaifaObserve = "TAIFA_OBSERVE"
	TargetSystemTaifaCitizen = "TAIFA_CITIZEN"
)

const (
	EffectAllow = "ALLOW"
	EffectDeny  = "DENY"
)

const (
	StatusActive   = "ACTIVE"
	StatusDisabled = "DISABLED"
)

const (
	DecisionAllow = "ALLOW"
	DecisionDeny  = "DENY"
)

const (
	DenyReasonMissingTargetSystem        = "missing_target_system"
	DenyReasonMissingMethod              = "missing_method"
	DenyReasonMissingRoute               = "missing_route"
	DenyReasonMissingOperation           = "missing_operation"
	DenyReasonPolicyNotFound             = "policy_not_found"
	DenyReasonPolicyDisabled             = "policy_disabled"
	DenyReasonExplicitPolicyDeny         = "explicit_policy_deny"
	DenyReasonMissingRequiredRole        = "missing_required_role"
	DenyReasonMissingRequiredCapability  = "missing_required_capability"
	DenyReasonUnsupportedPolicyEffect    = "unsupported_policy_effect"
	DenyReasonPolicyEvaluationIncomplete = "policy_evaluation_incomplete"
)

type Policy struct {
	ID                   string
	TargetSystem         string
	RoutePattern         string
	Method               string
	Operation            string
	Effect               string
	Status               string
	Description          string
	RequiredRoles        []string
	RequiredCapabilities []string
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

type LookupInput struct {
	TargetSystem string
	Method       string
	Route        string
	Operation    string
}

func (i LookupInput) Normalized() LookupInput {
	return LookupInput{
		TargetSystem: strings.ToUpper(strings.TrimSpace(i.TargetSystem)),
		Method:       strings.ToUpper(strings.TrimSpace(i.Method)),
		Route:        strings.TrimSpace(i.Route),
		Operation:    strings.TrimSpace(i.Operation),
	}
}

type AuthorizationInput struct {
	TargetSystem             string
	Method                   string
	Route                    string
	Operation                string
	ActorRoles               []string
	OrganizationCapabilities []string
}

func (i AuthorizationInput) LookupInput() LookupInput {
	return LookupInput{
		TargetSystem: i.TargetSystem,
		Method:       i.Method,
		Route:        i.Route,
		Operation:    i.Operation,
	}.Normalized()
}

type EvaluationInput struct {
	TargetSystem             string
	Method                   string
	Route                    string
	Operation                string
	Policy                   *Policy
	ActorRoles               []string
	OrganizationCapabilities []string
}

type EvaluationResult struct {
	Decision                 string
	MatchedPolicyID          string
	DenyReason               string
	TargetSystem             string
	Method                   string
	Route                    string
	Operation                string
	RequiredRoles            []string
	RequiredCapabilities     []string
	ActorRoles               []string
	OrganizationCapabilities []string
}

func NormalizeStringSet(values []string) map[string]struct{} {
	normalized := make(map[string]struct{}, len(values))

	for _, value := range values {
		value = strings.ToUpper(strings.TrimSpace(value))
		if value == "" {
			continue
		}

		normalized[value] = struct{}{}
	}

	return normalized
}

func NormalizeStringSlice(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	normalized := make([]string, 0, len(values))

	for _, value := range values {
		value = strings.ToUpper(strings.TrimSpace(value))
		if value == "" {
			continue
		}

		if _, exists := seen[value]; exists {
			continue
		}

		seen[value] = struct{}{}
		normalized = append(normalized, value)
	}

	return normalized
}
