package exchange

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"taifa-exchange/internal/audit"
	"taifa-exchange/internal/platform/postgres"
)

type Execer interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{
		pool: pool,
	}
}

func (r *Repository) CreateDecisionWithAudit(
	ctx context.Context,
	decision DecisionRecord,
	event audit.Event,
) error {
	if r == nil || r.pool == nil {
		return fmt.Errorf("exchange repository is not configured")
	}

	return postgres.InTx(ctx, r.pool, func(ctx context.Context, tx pgx.Tx) error {
		if err := insertDecision(ctx, tx, decision); err != nil {
			return err
		}

		if err := audit.Insert(ctx, tx, event); err != nil {
			return err
		}

		return nil
	})
}

func insertDecision(ctx context.Context, execer Execer, decision DecisionRecord) error {
	if execer == nil {
		return fmt.Errorf("exchange decision execer is nil")
	}

	actorRolesJSON, err := json.Marshal(decision.ActorRoles)
	if err != nil {
		return fmt.Errorf("marshal actor roles: %w", err)
	}

	organizationCapabilitiesJSON, err := json.Marshal(decision.OrganizationCapabilities)
	if err != nil {
		return fmt.Errorf("marshal organization capabilities: %w", err)
	}

	requiredRolesJSON, err := json.Marshal(decision.RequiredRoles)
	if err != nil {
		return fmt.Errorf("marshal required roles: %w", err)
	}

	requiredCapabilitiesJSON, err := json.Marshal(decision.RequiredCapabilities)
	if err != nil {
		return fmt.Errorf("marshal required capabilities: %w", err)
	}

	const query = `
		INSERT INTO exchange_authorization_decisions (
			id,
			decision,
			target_system,
			route,
			method,
			operation,
			organization_id,
			actor_context_id,
			person_id,
			credential_id,
			session_id,
			matched_policy_id,
			deny_reason,
			request_correlation_id,
			taifa_id_correlation_id,
			actor_roles_json,
			organization_capabilities_json,
			required_roles_json,
			required_capabilities_json,
			created_at
		)
		VALUES (
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			$7,
			$8,
			$9,
			$10,
			$11,
			$12,
			$13,
			$14,
			$15,
			$16::jsonb,
			$17::jsonb,
			$18::jsonb,
			$19::jsonb,
			$20
		)
	`

	_, err = execer.Exec(
		ctx,
		query,
		decision.ID,
		decision.Decision,
		decision.TargetSystem,
		decision.Route,
		decision.Method,
		decision.Operation,
		decision.OrganizationID,
		nullIfEmpty(decision.ActorContextID),
		nullIfEmpty(decision.PersonID),
		nullIfEmpty(decision.CredentialID),
		nullIfEmpty(decision.SessionID),
		nullIfEmpty(decision.MatchedPolicyID),
		nullIfEmpty(decision.DenyReason),
		nullIfEmpty(decision.RequestCorrelationID),
		nullIfEmpty(decision.TaifaIDCorrelationID),
		string(actorRolesJSON),
		string(organizationCapabilitiesJSON),
		string(requiredRolesJSON),
		string(requiredCapabilitiesJSON),
		decision.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert exchange authorization decision: %w", err)
	}

	return nil
}

func nullIfEmpty(value string) any {
	if value == "" {
		return nil
	}

	return value
}
