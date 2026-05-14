package policy

import "testing"

func TestEvaluatorAllowsWhenPolicyRequirementsAreSatisfied(t *testing.T) {
	evaluator := NewEvaluator()

	result := evaluator.Evaluate(EvaluationInput{
		TargetSystem: TargetSystemTaifaCare,
		Method:       "POST",
		Route:        "/api/v1/claims",
		Operation:    "care.claim.create",
		Policy: &Policy{
			ID:                   "POL-CARE-CLAIM-CREATE-CLINICIAN",
			TargetSystem:         TargetSystemTaifaCare,
			Method:               "POST",
			RoutePattern:         "/api/v1/claims",
			Operation:            "care.claim.create",
			Effect:               EffectAllow,
			Status:               StatusActive,
			RequiredRoles:        []string{"PROVIDER_CLINICIAN"},
			RequiredCapabilities: []string{"CAN_PROVIDE_HEALTH_SERVICES"},
		},
		ActorRoles:               []string{"PROVIDER_CLINICIAN"},
		OrganizationCapabilities: []string{"CAN_PROVIDE_HEALTH_SERVICES"},
	})

	if result.Decision != DecisionAllow {
		t.Fatalf("expected %s, got %s with deny reason %q", DecisionAllow, result.Decision, result.DenyReason)
	}

	if result.MatchedPolicyID != "POL-CARE-CLAIM-CREATE-CLINICIAN" {
		t.Fatalf("expected matched policy id, got %q", result.MatchedPolicyID)
	}
}

func TestEvaluatorDeniesWhenPolicyNotFound(t *testing.T) {
	evaluator := NewEvaluator()

	result := evaluator.Evaluate(EvaluationInput{
		TargetSystem: TargetSystemTaifaPay,
		Method:       "POST",
		Route:        "/api/v1/payment-instructions",
		Operation:    "pay.instruction.create",
		Policy:       nil,
		ActorRoles:   []string{"PROVIDER_CLINICIAN"},
	})

	if result.Decision != DecisionDeny {
		t.Fatalf("expected %s, got %s", DecisionDeny, result.Decision)
	}

	if result.DenyReason != DenyReasonPolicyNotFound {
		t.Fatalf("expected deny reason %q, got %q", DenyReasonPolicyNotFound, result.DenyReason)
	}
}

func TestEvaluatorDeniesWhenRequiredRoleIsMissing(t *testing.T) {
	evaluator := NewEvaluator()

	result := evaluator.Evaluate(EvaluationInput{
		TargetSystem: TargetSystemTaifaPay,
		Method:       "POST",
		Route:        "/api/v1/payment-instructions",
		Operation:    "pay.instruction.create",
		Policy: &Policy{
			ID:                   "POL-PAY-INSTRUCTION-CREATE-PAY-OPERATOR",
			Effect:               EffectAllow,
			Status:               StatusActive,
			RequiredRoles:        []string{"PAY_OPERATOR"},
			RequiredCapabilities: []string{"CAN_ROUTE_PAYMENTS"},
		},
		ActorRoles:               []string{"PROVIDER_CLINICIAN"},
		OrganizationCapabilities: []string{"CAN_ROUTE_PAYMENTS"},
	})

	if result.Decision != DecisionDeny {
		t.Fatalf("expected %s, got %s", DecisionDeny, result.Decision)
	}

	if result.DenyReason != DenyReasonMissingRequiredRole {
		t.Fatalf("expected deny reason %q, got %q", DenyReasonMissingRequiredRole, result.DenyReason)
	}
}

func TestEvaluatorDeniesWhenRequiredCapabilityIsMissing(t *testing.T) {
	evaluator := NewEvaluator()

	result := evaluator.Evaluate(EvaluationInput{
		TargetSystem: TargetSystemTaifaCare,
		Method:       "POST",
		Route:        "/api/v1/claims",
		Operation:    "care.claim.create",
		Policy: &Policy{
			ID:                   "POL-CARE-CLAIM-CREATE-CLINICIAN",
			Effect:               EffectAllow,
			Status:               StatusActive,
			RequiredRoles:        []string{"PROVIDER_CLINICIAN"},
			RequiredCapabilities: []string{"CAN_PROVIDE_HEALTH_SERVICES"},
		},
		ActorRoles:               []string{"PROVIDER_CLINICIAN"},
		OrganizationCapabilities: []string{"CAN_ROUTE_PAYMENTS"},
	})

	if result.Decision != DecisionDeny {
		t.Fatalf("expected %s, got %s", DecisionDeny, result.Decision)
	}

	if result.DenyReason != DenyReasonMissingRequiredCapability {
		t.Fatalf("expected deny reason %q, got %q", DenyReasonMissingRequiredCapability, result.DenyReason)
	}
}

func TestEvaluatorUsesAnyRoleSemantics(t *testing.T) {
	evaluator := NewEvaluator()

	result := evaluator.Evaluate(EvaluationInput{
		TargetSystem: TargetSystemTaifaObserve,
		Method:       "GET",
		Route:        "/api/v1/security-events",
		Operation:    "observe.security_event.read",
		Policy: &Policy{
			ID:                   "POL-OBSERVE-SECURITY-EVENT-READ",
			Effect:               EffectAllow,
			Status:               StatusActive,
			RequiredRoles:        []string{"OBSERVE_ANALYST", "OBSERVE_AUDITOR"},
			RequiredCapabilities: []string{"CAN_OBSERVE_SECURITY_EVENTS"},
		},
		ActorRoles:               []string{"OBSERVE_AUDITOR"},
		OrganizationCapabilities: []string{"CAN_OBSERVE_SECURITY_EVENTS"},
	})

	if result.Decision != DecisionAllow {
		t.Fatalf("expected %s, got %s with deny reason %q", DecisionAllow, result.Decision, result.DenyReason)
	}
}

func TestEvaluatorDeniesExplicitDenyPolicy(t *testing.T) {
	evaluator := NewEvaluator()

	result := evaluator.Evaluate(EvaluationInput{
		TargetSystem: TargetSystemTaifaCare,
		Method:       "POST",
		Route:        "/api/v1/claims",
		Operation:    "care.claim.create",
		Policy: &Policy{
			ID:            "POL-DENY-EXAMPLE",
			Effect:        EffectDeny,
			Status:        StatusActive,
			RequiredRoles: []string{"PROVIDER_CLINICIAN"},
		},
		ActorRoles: []string{"PROVIDER_CLINICIAN"},
	})

	if result.Decision != DecisionDeny {
		t.Fatalf("expected %s, got %s", DecisionDeny, result.Decision)
	}

	if result.DenyReason != DenyReasonExplicitPolicyDeny {
		t.Fatalf("expected deny reason %q, got %q", DenyReasonExplicitPolicyDeny, result.DenyReason)
	}
}
