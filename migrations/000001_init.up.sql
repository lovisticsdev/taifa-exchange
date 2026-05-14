CREATE TABLE IF NOT EXISTS exchange_policies (
	id TEXT PRIMARY KEY,
	target_system TEXT NOT NULL,
	route_pattern TEXT NOT NULL,
	method TEXT NOT NULL,
	operation TEXT NOT NULL,
	effect TEXT NOT NULL,
	status TEXT NOT NULL DEFAULT 'ACTIVE',
	description TEXT NOT NULL DEFAULT '',
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

	CONSTRAINT exchange_policies_target_system_check
		CHECK (
			target_system IN (
				'TAIFA_ID',
				'TAIFA_CARE',
				'TAIFA_TAX',
				'TAIFA_PAY',
				'TAIFA_OBSERVE',
				'TAIFA_CITIZEN'
			)
		),

	CONSTRAINT exchange_policies_method_check
		CHECK (
			method IN (
				'GET',
				'POST',
				'PUT',
				'PATCH',
				'DELETE'
			)
		),

	CONSTRAINT exchange_policies_effect_check
		CHECK (
			effect IN (
				'ALLOW',
				'DENY'
			)
		),

	CONSTRAINT exchange_policies_status_check
		CHECK (
			status IN (
				'ACTIVE',
				'DISABLED'
			)
		),

	CONSTRAINT exchange_policies_route_pattern_not_blank_check
		CHECK (length(trim(route_pattern)) > 0),

	CONSTRAINT exchange_policies_operation_not_blank_check
		CHECK (length(trim(operation)) > 0)
);

CREATE UNIQUE INDEX IF NOT EXISTS exchange_policies_unique_route_operation_idx
	ON exchange_policies (
		target_system,
		method,
		route_pattern,
		operation
	);

CREATE TABLE IF NOT EXISTS exchange_policy_roles (
	id TEXT PRIMARY KEY,
	policy_id TEXT NOT NULL REFERENCES exchange_policies(id) ON DELETE CASCADE,
	role TEXT NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

	CONSTRAINT exchange_policy_roles_role_check
		CHECK (
			role IN (
				'CITIZEN',
				'PROVIDER_CLINICIAN',
				'PROVIDER_CLAIMS_OFFICER',
				'CARE_ADJUDICATOR',
				'TAX_OFFICER',
				'EMPLOYER_SUBMITTER',
				'PAY_OPERATOR',
				'OBSERVE_ANALYST',
				'OBSERVE_AUDITOR',
				'SYSTEM_ADMIN'
			)
		)
);

CREATE UNIQUE INDEX IF NOT EXISTS exchange_policy_roles_unique_idx
	ON exchange_policy_roles (
		policy_id,
		role
	);

CREATE TABLE IF NOT EXISTS exchange_policy_capabilities (
	id TEXT PRIMARY KEY,
	policy_id TEXT NOT NULL REFERENCES exchange_policies(id) ON DELETE CASCADE,
	capability TEXT NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

	CONSTRAINT exchange_policy_capabilities_capability_check
		CHECK (
			capability IN (
				'CAN_EMPLOY_PERSONS',
				'CAN_SUBMIT_TAX_CONTRIBUTIONS',
				'CAN_PROVIDE_HEALTH_SERVICES',
				'CAN_RECEIVE_HEALTH_PAYOUTS',
				'CAN_ROUTE_PAYMENTS',
				'CAN_HOLD_RESERVE_ACCOUNT',
				'CAN_OPERATE_GOVERNMENT_SERVICE',
				'CAN_OBSERVE_SECURITY_EVENTS'
			)
		)
);

CREATE UNIQUE INDEX IF NOT EXISTS exchange_policy_capabilities_unique_idx
	ON exchange_policy_capabilities (
		policy_id,
		capability
	);

CREATE TABLE IF NOT EXISTS exchange_authorization_decisions (
	id TEXT PRIMARY KEY,
	decision TEXT NOT NULL,
	target_system TEXT NOT NULL,
	route TEXT NOT NULL,
	method TEXT NOT NULL,
	operation TEXT NOT NULL,
	organization_id TEXT NOT NULL,
	actor_context_id TEXT,
	person_id TEXT,
	credential_id TEXT,
	session_id TEXT,
	matched_policy_id TEXT REFERENCES exchange_policies(id),
	deny_reason TEXT,
	request_correlation_id TEXT,
	taifa_id_correlation_id TEXT,
	actor_roles_json JSONB NOT NULL DEFAULT '[]'::jsonb,
	organization_capabilities_json JSONB NOT NULL DEFAULT '[]'::jsonb,
	required_roles_json JSONB NOT NULL DEFAULT '[]'::jsonb,
	required_capabilities_json JSONB NOT NULL DEFAULT '[]'::jsonb,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

	CONSTRAINT exchange_authorization_decisions_decision_check
		CHECK (
			decision IN (
				'ALLOW',
				'DENY'
			)
		),

	CONSTRAINT exchange_authorization_decisions_target_system_check
		CHECK (
			target_system IN (
				'TAIFA_ID',
				'TAIFA_CARE',
				'TAIFA_TAX',
				'TAIFA_PAY',
				'TAIFA_OBSERVE',
				'TAIFA_CITIZEN'
			)
		),

	CONSTRAINT exchange_authorization_decisions_method_check
		CHECK (
			method IN (
				'GET',
				'POST',
				'PUT',
				'PATCH',
				'DELETE'
			)
		)
);

CREATE INDEX IF NOT EXISTS exchange_authorization_decisions_created_at_idx
	ON exchange_authorization_decisions (created_at DESC);

CREATE INDEX IF NOT EXISTS exchange_authorization_decisions_actor_idx
	ON exchange_authorization_decisions (
		person_id,
		organization_id,
		created_at DESC
	);

CREATE INDEX IF NOT EXISTS exchange_authorization_decisions_route_idx
	ON exchange_authorization_decisions (
		target_system,
		method,
		route,
		operation,
		created_at DESC
	);

CREATE TABLE IF NOT EXISTS audit_outbox (
	id TEXT PRIMARY KEY,
	event_type TEXT NOT NULL,
	source_system TEXT NOT NULL,
	actor_id TEXT,
	subject_id TEXT,
	resource_type TEXT NOT NULL,
	resource_id TEXT NOT NULL,
	action TEXT NOT NULL,
	result TEXT NOT NULL,
	correlation_id TEXT,
	payload_json JSONB NOT NULL DEFAULT '{}'::jsonb,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	published_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS audit_outbox_created_at_idx
	ON audit_outbox (created_at DESC);

CREATE INDEX IF NOT EXISTS audit_outbox_unpublished_idx
	ON audit_outbox (created_at ASC)
	WHERE published_at IS NULL;

CREATE INDEX IF NOT EXISTS audit_outbox_event_type_idx
	ON audit_outbox (event_type, created_at DESC);

CREATE INDEX IF NOT EXISTS audit_outbox_correlation_idx
	ON audit_outbox (correlation_id, created_at DESC);