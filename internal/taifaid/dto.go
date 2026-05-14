package taifaid

import "encoding/json"

type ReadyResponse struct {
	CorrelationID string            `json:"correlation_id"`
	Status        string            `json:"status"`
	Service       string            `json:"service"`
	Environment   string            `json:"environment,omitempty"`
	Dependencies  map[string]string `json:"dependencies,omitempty"`
}

type ErrorEnvelope struct {
	CorrelationID string    `json:"correlation_id"`
	Error         ErrorBody `json:"error"`
}

type ErrorBody struct {
	Code          string `json:"code"`
	CorrelationID string `json:"correlation_id"`
	Message       string `json:"message"`
}

type DataEnvelope struct {
	CorrelationID string          `json:"correlation_id"`
	Data          json.RawMessage `json:"data"`
}

type ResolveActorContextRequest struct {
	Token          string `json:"token"`
	OrganizationID string `json:"organization_id"`
}

type ActorContextResponse struct {
	CorrelationID string       `json:"correlation_id"`
	Data          ActorContext `json:"data"`
}

type ActorContext struct {
	ActorContextID     string            `json:"actor_context_id"`
	PersonID           string            `json:"person_id"`
	CredentialID       string            `json:"credential_id"`
	Username           string            `json:"username,omitempty"`
	OrganizationID     string            `json:"organization_id"`
	OrganizationName   string            `json:"organization_name,omitempty"`
	OrganizationType   string            `json:"organization_type,omitempty"`
	OrganizationStatus string            `json:"organization_status,omitempty"`
	SessionID          string            `json:"session_id"`
	Roles              []string          `json:"roles"`
	Memberships        []ActorMembership `json:"memberships"`
}

type ActorMembership struct {
	ID             string   `json:"id"`
	MembershipID   string   `json:"membership_id,omitempty"`
	MembershipType string   `json:"membership_type"`
	Status         string   `json:"status,omitempty"`
	Roles          []string `json:"roles,omitempty"`
}

type CapabilitiesResponse struct {
	CorrelationID string                   `json:"correlation_id"`
	Data          []OrganizationCapability `json:"data"`
}

type OrganizationCapability struct {
	ID             string `json:"id,omitempty"`
	OrganizationID string `json:"organization_id,omitempty"`
	Capability     string `json:"capability"`
	CreatedAt      string `json:"created_at,omitempty"`
}
