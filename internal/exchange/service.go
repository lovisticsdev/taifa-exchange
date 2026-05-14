package exchange

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"taifa-exchange/internal/audit"
	"taifa-exchange/internal/platform/ids"
	"taifa-exchange/internal/policy"
	"taifa-exchange/internal/taifaid"
)

type Service struct {
	repository    *Repository
	policyService *policy.Service
	taifaIDClient *taifaid.Client
}

func NewService(
	repository *Repository,
	policyService *policy.Service,
	taifaIDClient *taifaid.Client,
) *Service {
	return &Service{
		repository:    repository,
		policyService: policyService,
		taifaIDClient: taifaIDClient,
	}
}

func (s *Service) Authorize(ctx context.Context, input AuthorizeInput) (AuthorizeResult, error) {
	request := normalizeAuthorizeRequest(input.Request)
	correlationID := strings.TrimSpace(input.CorrelationID)
	token := strings.TrimSpace(input.Token)

	decision := DecisionRecord{
		ID:                   ids.NewDecisionID(),
		Decision:             DecisionDeny,
		TargetSystem:         NormalizeTargetSystemForDecision(request.TargetSystem),
		Route:                NormalizeRoute(request.Route),
		Method:               NormalizeMethodForDecision(request.Method),
		Operation:            NormalizeOperation(request.Operation),
		OrganizationID:       NormalizeOrganizationID(request.OrganizationID),
		RequestCorrelationID: correlationID,
		CreatedAt:            time.Now().UTC(),
	}

	if denyReason, statusCode, errorCode := validateRequestForAuthorization(request, token); denyReason != "" {
		decision.DenyReason = denyReason

		result := denyResult(
			decision,
			statusCode,
			errorCode,
			messageForStatus(statusCode),
		)

		if err := s.persist(ctx, result); err != nil {
			return AuthorizeResult{}, err
		}

		return result, nil
	}

	if s == nil || s.taifaIDClient == nil || !s.taifaIDClient.IsConfigured() {
		decision.DenyReason = DenyReasonTaifaIDClientNotConfigured

		result := denyResult(
			decision,
			http.StatusServiceUnavailable,
			ErrorCodeUnavailable,
			"Exchange upstream identity dependency is unavailable.",
		)

		if err := s.persist(ctx, result); err != nil {
			return AuthorizeResult{}, err
		}

		return result, nil
	}

	actorResponse, err := s.taifaIDClient.ResolveActorContext(
		ctx,
		token,
		request.OrganizationID,
		correlationID,
	)
	if err != nil {
		decision.DenyReason = DenyReasonActorContextResolveFailed

		result := denyResult(
			decision,
			http.StatusUnauthorized,
			ErrorCodeUnauthorized,
			"Exchange could not validate the actor context.",
		)

		if err := s.persist(ctx, result); err != nil {
			return AuthorizeResult{}, err
		}

		return result, nil
	}

	actorContext := actorResponse.Data
	decision.TaifaIDCorrelationID = actorResponse.CorrelationID
	decision.ActorContextID = actorContext.ActorContextID
	decision.PersonID = actorContext.PersonID
	decision.CredentialID = actorContext.CredentialID
	decision.SessionID = actorContext.SessionID
	decision.ActorRoles = collectActorRoles(actorContext)

	capabilitiesResponse, err := s.taifaIDClient.ListOrganizationCapabilities(
		ctx,
		request.OrganizationID,
		correlationID,
	)
	if err != nil {
		decision.DenyReason = DenyReasonOrganizationCapabilitiesLookupFailed

		result := denyResult(
			decision,
			http.StatusBadGateway,
			ErrorCodeUnavailable,
			"Exchange could not validate organization capabilities.",
		)

		if err := s.persist(ctx, result); err != nil {
			return AuthorizeResult{}, err
		}

		return result, nil
	}

	decision.OrganizationCapabilities = collectCapabilities(capabilitiesResponse.Data)

	if s.policyService == nil {
		decision.DenyReason = DenyReasonPolicyEvaluationFailed

		result := denyResult(
			decision,
			http.StatusServiceUnavailable,
			ErrorCodeUnavailable,
			"Exchange policy service is unavailable.",
		)

		if err := s.persist(ctx, result); err != nil {
			return AuthorizeResult{}, err
		}

		return result, nil
	}

	evaluation, err := s.policyService.Evaluate(ctx, policy.AuthorizationInput{
		TargetSystem:             request.TargetSystem,
		Method:                   request.Method,
		Route:                    request.Route,
		Operation:                request.Operation,
		ActorRoles:               decision.ActorRoles,
		OrganizationCapabilities: decision.OrganizationCapabilities,
	})
	if err != nil {
		return AuthorizeResult{}, fmt.Errorf("evaluate exchange policy: %w", err)
	}

	decision.Decision = evaluation.Decision
	decision.MatchedPolicyID = evaluation.MatchedPolicyID
	decision.DenyReason = evaluation.DenyReason
	decision.RequiredRoles = evaluation.RequiredRoles
	decision.RequiredCapabilities = evaluation.RequiredCapabilities

	if evaluation.Decision == DecisionAllow {
		result := allowResult(decision, actorContext)

		if err := s.persist(ctx, result); err != nil {
			return AuthorizeResult{}, err
		}

		return result, nil
	}

	result := denyResult(
		decision,
		http.StatusForbidden,
		ErrorCodeForbidden,
		"Exchange policy denied the request.",
	)

	if err := s.persist(ctx, result); err != nil {
		return AuthorizeResult{}, err
	}

	return result, nil
}

func (s *Service) persist(ctx context.Context, result AuthorizeResult) error {
	if s == nil || s.repository == nil {
		return fmt.Errorf("exchange service repository is not configured")
	}

	event := auditEventForResult(result)

	if err := s.repository.CreateDecisionWithAudit(ctx, result.DecisionRecord, event); err != nil {
		return fmt.Errorf("persist exchange authorization decision: %w", err)
	}

	return nil
}

func normalizeAuthorizeRequest(request AuthorizeRequest) AuthorizeRequest {
	return AuthorizeRequest{
		OrganizationID: strings.TrimSpace(request.OrganizationID),
		TargetSystem:   strings.ToUpper(strings.TrimSpace(request.TargetSystem)),
		Route:          strings.TrimSpace(request.Route),
		Method:         strings.ToUpper(strings.TrimSpace(request.Method)),
		Operation:      strings.TrimSpace(request.Operation),
	}
}

func validateRequestForAuthorization(
	request AuthorizeRequest,
	token string,
) (denyReason string, statusCode int, errorCode string) {
	if strings.TrimSpace(token) == "" {
		return DenyReasonMissingToken, http.StatusUnauthorized, ErrorCodeUnauthorized
	}

	if strings.TrimSpace(request.OrganizationID) == "" {
		return DenyReasonMissingOrganizationID, http.StatusBadRequest, ErrorCodeValidation
	}

	if strings.TrimSpace(request.TargetSystem) == "" {
		return DenyReasonMissingTargetSystem, http.StatusBadRequest, ErrorCodeValidation
	}

	if !IsKnownTargetSystem(request.TargetSystem) {
		return DenyReasonUnsupportedTargetSystem, http.StatusBadRequest, ErrorCodeValidation
	}

	if strings.TrimSpace(request.Method) == "" {
		return DenyReasonMissingMethod, http.StatusBadRequest, ErrorCodeValidation
	}

	if !IsKnownMethod(request.Method) {
		return DenyReasonUnsupportedMethod, http.StatusBadRequest, ErrorCodeValidation
	}

	if strings.TrimSpace(request.Route) == "" {
		return DenyReasonMissingRoute, http.StatusBadRequest, ErrorCodeValidation
	}

	if strings.TrimSpace(request.Operation) == "" {
		return DenyReasonMissingOperation, http.StatusBadRequest, ErrorCodeValidation
	}

	return "", 0, ""
}

func collectActorRoles(actorContext taifaid.ActorContext) []string {
	values := make([]string, 0)

	values = append(values, actorContext.Roles...)

	for _, membership := range actorContext.Memberships {
		values = append(values, membership.Roles...)
	}

	return normalizeStringSlice(values)
}

func collectCapabilities(capabilities []taifaid.OrganizationCapability) []string {
	values := make([]string, 0, len(capabilities))

	for _, capability := range capabilities {
		values = append(values, capability.Capability)
	}

	return normalizeStringSlice(values)
}

func normalizeStringSlice(values []string) []string {
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

func allowResult(decision DecisionRecord, actorContext taifaid.ActorContext) AuthorizeResult {
	response := &AuthorizeResponse{
		DecisionID:      decision.ID,
		Decision:        decision.Decision,
		TargetSystem:    decision.TargetSystem,
		Route:           decision.Route,
		Method:          decision.Method,
		Operation:       decision.Operation,
		MatchedPolicyID: decision.MatchedPolicyID,
		ActorContext: ActorContextSummary{
			ActorContextID: decision.ActorContextID,
			PersonID:       decision.PersonID,
			CredentialID:   decision.CredentialID,
			OrganizationID: decision.OrganizationID,
			SessionID:      decision.SessionID,
			Roles:          decision.ActorRoles,
			Memberships:    summarizeMemberships(actorContext.Memberships),
		},
		Obligations: Obligations{
			PropagateCorrelationID: true,
			RequireAudit:           true,
		},
	}

	return AuthorizeResult{
		DecisionRecord: decision,
		Response:       response,
		HTTPStatusCode: http.StatusOK,
	}
}

func denyResult(
	decision DecisionRecord,
	statusCode int,
	errorCode string,
	errorMessage string,
) AuthorizeResult {
	if statusCode == 0 {
		statusCode = http.StatusForbidden
	}

	if errorCode == "" {
		errorCode = ErrorCodeForbidden
	}

	if errorMessage == "" {
		errorMessage = "Exchange policy denied the request."
	}

	decision.Decision = DecisionDeny

	return AuthorizeResult{
		DecisionRecord: decision,
		HTTPStatusCode: statusCode,
		ErrorCode:      errorCode,
		ErrorMessage:   errorMessage,
	}
}

func summarizeMemberships(memberships []taifaid.ActorMembership) []MembershipSummary {
	summary := make([]MembershipSummary, 0, len(memberships))

	for _, membership := range memberships {
		id := strings.TrimSpace(membership.MembershipID)
		if id == "" {
			id = strings.TrimSpace(membership.ID)
		}

		summary = append(summary, MembershipSummary{
			ID:             id,
			MembershipType: membership.MembershipType,
			Roles:          normalizeStringSlice(membership.Roles),
		})
	}

	return summary
}

func messageForStatus(statusCode int) string {
	switch statusCode {
	case http.StatusBadRequest:
		return "Exchange authorization request was invalid."
	case http.StatusUnauthorized:
		return "Exchange authorization requires a valid bearer token."
	case http.StatusServiceUnavailable:
		return "Exchange dependency is unavailable."
	default:
		return "Exchange policy denied the request."
	}
}

func auditEventForResult(result AuthorizeResult) audit.Event {
	decision := result.DecisionRecord

	eventType := audit.EventAuthorizationDenied
	auditResult := audit.ResultDenied

	if decision.Decision == DecisionAllow {
		eventType = audit.EventAuthorizationAllowed
		auditResult = audit.ResultAllowed
	} else {
		switch decision.DenyReason {
		case DenyReasonActorContextResolveFailed:
			eventType = audit.EventActorContextResolveFailed
		case policy.DenyReasonPolicyNotFound:
			eventType = audit.EventPolicyNotFound
		}
	}

	return audit.Event{
		ID:            ids.NewAuditEventID(),
		EventType:     eventType,
		SourceSystem:  audit.SourceTaifaExchange,
		ActorID:       decision.PersonID,
		SubjectID:     decision.OrganizationID,
		ResourceType:  audit.ResourceAuthorizationDecision,
		ResourceID:    decision.ID,
		Action:        audit.ActionAuthorizeRequest,
		Result:        auditResult,
		CorrelationID: decision.RequestCorrelationID,
		Payload: map[string]any{
			"decision_id":               decision.ID,
			"decision":                  decision.Decision,
			"target_system":             decision.TargetSystem,
			"route":                     decision.Route,
			"method":                    decision.Method,
			"operation":                 decision.Operation,
			"organization_id":           decision.OrganizationID,
			"actor_context_id":          decision.ActorContextID,
			"person_id":                 decision.PersonID,
			"credential_id":             decision.CredentialID,
			"session_id":                decision.SessionID,
			"roles":                     decision.ActorRoles,
			"organization_capabilities": decision.OrganizationCapabilities,
			"matched_policy_id":         decision.MatchedPolicyID,
			"required_roles":            decision.RequiredRoles,
			"required_capabilities":     decision.RequiredCapabilities,
			"deny_reason":               decision.DenyReason,
			"taifa_id_correlation_id":   decision.TaifaIDCorrelationID,
			"exchange_correlation_id":   decision.RequestCorrelationID,
		},
		CreatedAt: decision.CreatedAt,
	}
}
