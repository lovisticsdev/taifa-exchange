package exchange

type AuthorizeRequest struct {
	OrganizationID string `json:"organization_id"`
	TargetSystem   string `json:"target_system"`
	Route          string `json:"route"`
	Method         string `json:"method"`
	Operation      string `json:"operation"`
}

type AuthorizeResponse struct {
	DecisionID      string              `json:"decision_id"`
	Decision        string              `json:"decision"`
	TargetSystem    string              `json:"target_system"`
	Route           string              `json:"route"`
	Method          string              `json:"method"`
	Operation       string              `json:"operation"`
	ActorContext    ActorContextSummary `json:"actor_context"`
	MatchedPolicyID string              `json:"matched_policy_id,omitempty"`
	Obligations     Obligations         `json:"obligations"`
}

type ActorContextSummary struct {
	ActorContextID string              `json:"actor_context_id,omitempty"`
	PersonID       string              `json:"person_id,omitempty"`
	CredentialID   string              `json:"credential_id,omitempty"`
	OrganizationID string              `json:"organization_id"`
	SessionID      string              `json:"session_id,omitempty"`
	Roles          []string            `json:"roles"`
	Memberships    []MembershipSummary `json:"memberships"`
}

type MembershipSummary struct {
	ID             string   `json:"id"`
	MembershipType string   `json:"membership_type"`
	Roles          []string `json:"roles,omitempty"`
}

type Obligations struct {
	PropagateCorrelationID bool `json:"propagate_correlation_id"`
	RequireAudit           bool `json:"require_audit"`
}
