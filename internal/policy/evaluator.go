package policy

import "strings"

type Evaluator struct{}

func NewEvaluator() Evaluator {
	return Evaluator{}
}

func (Evaluator) Evaluate(input EvaluationInput) EvaluationResult {
	targetSystem := strings.ToUpper(strings.TrimSpace(input.TargetSystem))
	method := strings.ToUpper(strings.TrimSpace(input.Method))
	route := strings.TrimSpace(input.Route)
	operation := strings.TrimSpace(input.Operation)

	result := EvaluationResult{
		Decision:                 DecisionDeny,
		TargetSystem:             targetSystem,
		Method:                   method,
		Route:                    route,
		Operation:                operation,
		ActorRoles:               NormalizeStringSlice(input.ActorRoles),
		OrganizationCapabilities: NormalizeStringSlice(input.OrganizationCapabilities),
	}

	if targetSystem == "" {
		result.DenyReason = DenyReasonMissingTargetSystem
		return result
	}

	if method == "" {
		result.DenyReason = DenyReasonMissingMethod
		return result
	}

	if route == "" {
		result.DenyReason = DenyReasonMissingRoute
		return result
	}

	if operation == "" {
		result.DenyReason = DenyReasonMissingOperation
		return result
	}

	if input.Policy == nil {
		result.DenyReason = DenyReasonPolicyNotFound
		return result
	}

	policy := *input.Policy

	result.MatchedPolicyID = policy.ID
	result.RequiredRoles = NormalizeStringSlice(policy.RequiredRoles)
	result.RequiredCapabilities = NormalizeStringSlice(policy.RequiredCapabilities)

	if strings.ToUpper(strings.TrimSpace(policy.Status)) != StatusActive {
		result.DenyReason = DenyReasonPolicyDisabled
		return result
	}

	switch strings.ToUpper(strings.TrimSpace(policy.Effect)) {
	case EffectDeny:
		result.DenyReason = DenyReasonExplicitPolicyDeny
		return result

	case EffectAllow:
		if len(result.RequiredRoles) > 0 && !hasAnyRequiredRole(result.ActorRoles, result.RequiredRoles) {
			result.DenyReason = DenyReasonMissingRequiredRole
			return result
		}

		if missing := missingRequiredCapabilities(result.OrganizationCapabilities, result.RequiredCapabilities); len(missing) > 0 {
			result.DenyReason = DenyReasonMissingRequiredCapability
			return result
		}

		result.Decision = DecisionAllow
		result.DenyReason = ""
		return result

	default:
		result.DenyReason = DenyReasonUnsupportedPolicyEffect
		return result
	}
}

func hasAnyRequiredRole(actorRoles []string, requiredRoles []string) bool {
	if len(requiredRoles) == 0 {
		return true
	}

	actorRoleSet := NormalizeStringSet(actorRoles)

	for _, requiredRole := range requiredRoles {
		requiredRole = strings.ToUpper(strings.TrimSpace(requiredRole))
		if requiredRole == "" {
			continue
		}

		if _, exists := actorRoleSet[requiredRole]; exists {
			return true
		}
	}

	return false
}

func missingRequiredCapabilities(
	organizationCapabilities []string,
	requiredCapabilities []string,
) []string {
	if len(requiredCapabilities) == 0 {
		return nil
	}

	capabilitySet := NormalizeStringSet(organizationCapabilities)
	missing := make([]string, 0)

	for _, requiredCapability := range requiredCapabilities {
		requiredCapability = strings.ToUpper(strings.TrimSpace(requiredCapability))
		if requiredCapability == "" {
			continue
		}

		if _, exists := capabilitySet[requiredCapability]; !exists {
			missing = append(missing, requiredCapability)
		}
	}

	return missing
}
