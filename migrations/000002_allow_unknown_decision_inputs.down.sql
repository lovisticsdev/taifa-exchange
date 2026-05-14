ALTER TABLE exchange_authorization_decisions
	DROP CONSTRAINT IF EXISTS exchange_authorization_decisions_target_system_check;

ALTER TABLE exchange_authorization_decisions
	ADD CONSTRAINT exchange_authorization_decisions_target_system_check
	CHECK (
		target_system IN (
			'TAIFA_ID',
			'TAIFA_CARE',
			'TAIFA_TAX',
			'TAIFA_PAY',
			'TAIFA_OBSERVE',
			'TAIFA_CITIZEN'
		)
	);

ALTER TABLE exchange_authorization_decisions
	DROP CONSTRAINT IF EXISTS exchange_authorization_decisions_method_check;

ALTER TABLE exchange_authorization_decisions
	ADD CONSTRAINT exchange_authorization_decisions_method_check
	CHECK (
		method IN (
			'GET',
			'POST',
			'PUT',
			'PATCH',
			'DELETE'
		)
	);