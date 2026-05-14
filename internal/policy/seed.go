package policy

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"taifa-exchange/internal/audit"
	"taifa-exchange/internal/platform/postgres"
)

type SeedPolicy struct {
	ID                   string
	TargetSystem         string
	RoutePattern         string
	Method               string
	Operation            string
	Effect               string
	Status               string
	Description          string
	RequiredRoles        []SeedPolicyRole
	RequiredCapabilities []SeedPolicyCapability
}

type SeedPolicyRole struct {
	ID   string
	Role string
}

type SeedPolicyCapability struct {
	ID         string
	Capability string
}

type SeedSummary struct {
	Policies     int
	Roles        int
	Capabilities int
	AuditEvents  int
}

type Seeder struct {
	pool *pgxpool.Pool
}

func NewSeeder(pool *pgxpool.Pool) *Seeder {
	return &Seeder{
		pool: pool,
	}
}

func (s *Seeder) SeedCanonical(ctx context.Context) (SeedSummary, error) {
	if s == nil || s.pool == nil {
		return SeedSummary{}, fmt.Errorf("policy seeder is not configured")
	}

	summary := SeedSummary{}

	err := postgres.InTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		for _, seed := range CanonicalPolicies() {
			insertedPolicy, err := insertSeedPolicy(ctx, tx, seed)
			if err != nil {
				return err
			}

			if insertedPolicy {
				summary.Policies++

				if err := audit.Insert(ctx, tx, policySeededAuditEvent(seed)); err != nil {
					return err
				}

				summary.AuditEvents++
			}

			for _, role := range seed.RequiredRoles {
				insertedRole, err := insertSeedPolicyRole(ctx, tx, seed.ID, role)
				if err != nil {
					return err
				}

				if insertedRole {
					summary.Roles++
				}
			}

			for _, capability := range seed.RequiredCapabilities {
				insertedCapability, err := insertSeedPolicyCapability(ctx, tx, seed.ID, capability)
				if err != nil {
					return err
				}

				if insertedCapability {
					summary.Capabilities++
				}
			}
		}

		return nil
	})
	if err != nil {
		return SeedSummary{}, err
	}

	return summary, nil
}

func CanonicalPolicies() []SeedPolicy {
	return []SeedPolicy{
		{
			ID:           "POL-CARE-CLAIM-CREATE-CLINICIAN",
			TargetSystem: TargetSystemTaifaCare,
			RoutePattern: "/api/v1/claims",
			Method:       "POST",
			Operation:    "care.claim.create",
			Effect:       EffectAllow,
			Status:       StatusActive,
			Description:  "Allow clinicians at health-service organizations to create care claims.",
			RequiredRoles: []SeedPolicyRole{
				{
					ID:   "PR-CARE-CLAIM-CREATE-CLINICIAN",
					Role: "PROVIDER_CLINICIAN",
				},
			},
			RequiredCapabilities: []SeedPolicyCapability{
				{
					ID:         "PC-CARE-CLAIM-CREATE-HEALTH-SERVICES",
					Capability: "CAN_PROVIDE_HEALTH_SERVICES",
				},
			},
		},
		{
			ID:           "POL-CARE-CLAIM-SUBMIT-CLAIMS-OFFICER",
			TargetSystem: TargetSystemTaifaCare,
			RoutePattern: "/api/v1/claims/submit",
			Method:       "POST",
			Operation:    "care.claim.submit",
			Effect:       EffectAllow,
			Status:       StatusActive,
			Description:  "Allow provider claims officers at health-service organizations to submit care claims.",
			RequiredRoles: []SeedPolicyRole{
				{
					ID:   "PR-CARE-CLAIM-SUBMIT-CLAIMS-OFFICER",
					Role: "PROVIDER_CLAIMS_OFFICER",
				},
			},
			RequiredCapabilities: []SeedPolicyCapability{
				{
					ID:         "PC-CARE-CLAIM-SUBMIT-HEALTH-SERVICES",
					Capability: "CAN_PROVIDE_HEALTH_SERVICES",
				},
			},
		},
		{
			ID:           "POL-TAX-CONTRIBUTION-SUBMIT-EMPLOYER",
			TargetSystem: TargetSystemTaifaTax,
			RoutePattern: "/api/v1/contributions",
			Method:       "POST",
			Operation:    "tax.contribution.submit",
			Effect:       EffectAllow,
			Status:       StatusActive,
			Description:  "Allow employer submitters at contribution-enabled organizations to submit tax contributions.",
			RequiredRoles: []SeedPolicyRole{
				{
					ID:   "PR-TAX-CONTRIBUTION-SUBMIT-EMPLOYER",
					Role: "EMPLOYER_SUBMITTER",
				},
			},
			RequiredCapabilities: []SeedPolicyCapability{
				{
					ID:         "PC-TAX-CONTRIBUTION-SUBMIT-CONTRIBUTIONS",
					Capability: "CAN_SUBMIT_TAX_CONTRIBUTIONS",
				},
			},
		},
		{
			ID:           "POL-TAX-CONTRIBUTION-READ-TAX-OFFICER",
			TargetSystem: TargetSystemTaifaTax,
			RoutePattern: "/api/v1/contributions",
			Method:       "GET",
			Operation:    "tax.contribution.read",
			Effect:       EffectAllow,
			Status:       StatusActive,
			Description:  "Allow tax officers at government-service organizations to read contribution records.",
			RequiredRoles: []SeedPolicyRole{
				{
					ID:   "PR-TAX-CONTRIBUTION-READ-TAX-OFFICER",
					Role: "TAX_OFFICER",
				},
			},
			RequiredCapabilities: []SeedPolicyCapability{
				{
					ID:         "PC-TAX-CONTRIBUTION-READ-GOVERNMENT-SERVICE",
					Capability: "CAN_OPERATE_GOVERNMENT_SERVICE",
				},
			},
		},
		{
			ID:           "POL-PAY-INSTRUCTION-CREATE-PAY-OPERATOR",
			TargetSystem: TargetSystemTaifaPay,
			RoutePattern: "/api/v1/payment-instructions",
			Method:       "POST",
			Operation:    "pay.instruction.create",
			Effect:       EffectAllow,
			Status:       StatusActive,
			Description:  "Allow payment operators at payment-routing organizations to create payment instructions.",
			RequiredRoles: []SeedPolicyRole{
				{
					ID:   "PR-PAY-INSTRUCTION-CREATE-PAY-OPERATOR",
					Role: "PAY_OPERATOR",
				},
			},
			RequiredCapabilities: []SeedPolicyCapability{
				{
					ID:         "PC-PAY-INSTRUCTION-CREATE-ROUTE-PAYMENTS",
					Capability: "CAN_ROUTE_PAYMENTS",
				},
			},
		},
		{
			ID:           "POL-OBSERVE-SECURITY-EVENT-READ-ANALYST",
			TargetSystem: TargetSystemTaifaObserve,
			RoutePattern: "/api/v1/security-events",
			Method:       "GET",
			Operation:    "observe.security_event.read",
			Effect:       EffectAllow,
			Status:       StatusActive,
			Description:  "Allow observe analysts at observe-enabled organizations to read security events.",
			RequiredRoles: []SeedPolicyRole{
				{
					ID:   "PR-OBSERVE-SECURITY-EVENT-READ-ANALYST",
					Role: "OBSERVE_ANALYST",
				},
			},
			RequiredCapabilities: []SeedPolicyCapability{
				{
					ID:         "PC-OBSERVE-SECURITY-EVENT-READ-OBSERVE-EVENTS",
					Capability: "CAN_OBSERVE_SECURITY_EVENTS",
				},
			},
		},
	}
}

func insertSeedPolicy(ctx context.Context, tx pgx.Tx, seed SeedPolicy) (bool, error) {
	const query = `
		INSERT INTO exchange_policies (
			id,
			target_system,
			route_pattern,
			method,
			operation,
			effect,
			status,
			description
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT DO NOTHING
	`

	tag, err := tx.Exec(
		ctx,
		query,
		seed.ID,
		seed.TargetSystem,
		seed.RoutePattern,
		seed.Method,
		seed.Operation,
		seed.Effect,
		seed.Status,
		seed.Description,
	)
	if err != nil {
		return false, fmt.Errorf("insert seed exchange policy %s: %w", seed.ID, err)
	}

	return tag.RowsAffected() > 0, nil
}

func insertSeedPolicyRole(
	ctx context.Context,
	tx pgx.Tx,
	policyID string,
	role SeedPolicyRole,
) (bool, error) {
	const query = `
		INSERT INTO exchange_policy_roles (
			id,
			policy_id,
			role
		)
		VALUES ($1, $2, $3)
		ON CONFLICT DO NOTHING
	`

	tag, err := tx.Exec(
		ctx,
		query,
		role.ID,
		policyID,
		role.Role,
	)
	if err != nil {
		return false, fmt.Errorf("insert seed exchange policy role %s: %w", role.ID, err)
	}

	return tag.RowsAffected() > 0, nil
}

func insertSeedPolicyCapability(
	ctx context.Context,
	tx pgx.Tx,
	policyID string,
	capability SeedPolicyCapability,
) (bool, error) {
	const query = `
		INSERT INTO exchange_policy_capabilities (
			id,
			policy_id,
			capability
		)
		VALUES ($1, $2, $3)
		ON CONFLICT DO NOTHING
	`

	tag, err := tx.Exec(
		ctx,
		query,
		capability.ID,
		policyID,
		capability.Capability,
	)
	if err != nil {
		return false, fmt.Errorf("insert seed exchange policy capability %s: %w", capability.ID, err)
	}

	return tag.RowsAffected() > 0, nil
}

func policySeededAuditEvent(seed SeedPolicy) audit.Event {
	requiredRoles := make([]string, 0, len(seed.RequiredRoles))
	for _, role := range seed.RequiredRoles {
		requiredRoles = append(requiredRoles, role.Role)
	}

	requiredCapabilities := make([]string, 0, len(seed.RequiredCapabilities))
	for _, capability := range seed.RequiredCapabilities {
		requiredCapabilities = append(requiredCapabilities, capability.Capability)
	}

	return audit.Event{
		EventType:    audit.EventPolicySeeded,
		SourceSystem: audit.SourceTaifaExchange,
		ResourceType: audit.ResourceExchangePolicy,
		ResourceID:   seed.ID,
		Action:       audit.ActionCreatePolicy,
		Result:       audit.ResultSuccess,
		Payload: map[string]any{
			"policy_id":             seed.ID,
			"target_system":         seed.TargetSystem,
			"route_pattern":         seed.RoutePattern,
			"method":                seed.Method,
			"operation":             seed.Operation,
			"effect":                seed.Effect,
			"status":                seed.Status,
			"required_roles":        requiredRoles,
			"required_capabilities": requiredCapabilities,
			"description":           seed.Description,
		},
		CreatedAt: time.Now().UTC(),
	}
}
