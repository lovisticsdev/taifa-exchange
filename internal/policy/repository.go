package policy

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrPolicyNotFound = errors.New("exchange policy not found")

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{
		pool: pool,
	}
}

func (r *Repository) FindActiveByRoute(ctx context.Context, input LookupInput) (*Policy, error) {
	if r == nil || r.pool == nil {
		return nil, fmt.Errorf("policy repository is not configured")
	}

	input = input.Normalized()

	const query = `
		SELECT
			id,
			target_system,
			route_pattern,
			method,
			operation,
			effect,
			status,
			description,
			created_at,
			updated_at
		FROM exchange_policies
		WHERE target_system = $1
		  AND method = $2
		  AND route_pattern = $3
		  AND operation = $4
		  AND status = 'ACTIVE'
		LIMIT 1
	`

	var found Policy

	err := r.pool.QueryRow(
		ctx,
		query,
		input.TargetSystem,
		input.Method,
		input.Route,
		input.Operation,
	).Scan(
		&found.ID,
		&found.TargetSystem,
		&found.RoutePattern,
		&found.Method,
		&found.Operation,
		&found.Effect,
		&found.Status,
		&found.Description,
		&found.CreatedAt,
		&found.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrPolicyNotFound
		}

		return nil, fmt.Errorf("find exchange policy: %w", err)
	}

	roles, err := r.listPolicyRoles(ctx, found.ID)
	if err != nil {
		return nil, err
	}

	capabilities, err := r.listPolicyCapabilities(ctx, found.ID)
	if err != nil {
		return nil, err
	}

	found.RequiredRoles = roles
	found.RequiredCapabilities = capabilities

	return &found, nil
}

func (r *Repository) listPolicyRoles(ctx context.Context, policyID string) ([]string, error) {
	const query = `
		SELECT role
		FROM exchange_policy_roles
		WHERE policy_id = $1
		ORDER BY role ASC
	`

	rows, err := r.pool.Query(ctx, query, policyID)
	if err != nil {
		return nil, fmt.Errorf("list exchange policy roles: %w", err)
	}
	defer rows.Close()

	roles := make([]string, 0)

	for rows.Next() {
		var role string
		if err := rows.Scan(&role); err != nil {
			return nil, fmt.Errorf("scan exchange policy role: %w", err)
		}

		roles = append(roles, role)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate exchange policy roles: %w", err)
	}

	return roles, nil
}

func (r *Repository) listPolicyCapabilities(ctx context.Context, policyID string) ([]string, error) {
	const query = `
		SELECT capability
		FROM exchange_policy_capabilities
		WHERE policy_id = $1
		ORDER BY capability ASC
	`

	rows, err := r.pool.Query(ctx, query, policyID)
	if err != nil {
		return nil, fmt.Errorf("list exchange policy capabilities: %w", err)
	}
	defer rows.Close()

	capabilities := make([]string, 0)

	for rows.Next() {
		var capability string
		if err := rows.Scan(&capability); err != nil {
			return nil, fmt.Errorf("scan exchange policy capability: %w", err)
		}

		capabilities = append(capabilities, capability)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate exchange policy capabilities: %w", err)
	}

	return capabilities, nil
}
