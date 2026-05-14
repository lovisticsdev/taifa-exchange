package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

const (
	defaultTaifaIDBaseURL       = "http://localhost:8080"
	defaultTaifaExchangeBaseURL = "http://localhost:8081"

	defaultUsername       = "clinician.seed"
	defaultPassword       = "ExampleDevPass123!"
	defaultOrganizationID = "ORG-HP-CLINIC"
)

type testConfig struct {
	TaifaIDBaseURL       string
	TaifaExchangeBaseURL string
	Username             string
	Password             string
	OrganizationID       string
	Force                bool
}

type errorEnvelope struct {
	CorrelationID string    `json:"correlation_id"`
	Error         errorBody `json:"error"`
}

type errorBody struct {
	Code          string `json:"code"`
	CorrelationID string `json:"correlation_id"`
	Message       string `json:"message"`
}

type loginEnvelope struct {
	CorrelationID string          `json:"correlation_id"`
	Data          json.RawMessage `json:"data"`
}

type authorizeEnvelope struct {
	CorrelationID string            `json:"correlation_id"`
	Data          authorizeResponse `json:"data"`
}

type authorizeResponse struct {
	DecisionID      string              `json:"decision_id"`
	Decision        string              `json:"decision"`
	TargetSystem    string              `json:"target_system"`
	Route           string              `json:"route"`
	Method          string              `json:"method"`
	Operation       string              `json:"operation"`
	ActorContext    actorContextSummary `json:"actor_context"`
	MatchedPolicyID string              `json:"matched_policy_id"`
	Obligations     obligations         `json:"obligations"`
}

type actorContextSummary struct {
	ActorContextID string              `json:"actor_context_id"`
	PersonID       string              `json:"person_id"`
	CredentialID   string              `json:"credential_id"`
	OrganizationID string              `json:"organization_id"`
	SessionID      string              `json:"session_id"`
	Roles          []string            `json:"roles"`
	Memberships    []membershipSummary `json:"memberships"`
}

type membershipSummary struct {
	ID             string   `json:"id"`
	MembershipType string   `json:"membership_type"`
	Roles          []string `json:"roles"`
}

type obligations struct {
	PropagateCorrelationID bool `json:"propagate_correlation_id"`
	RequireAudit           bool `json:"require_audit"`
}

func TestTaifaExchangeAuthorizationSmoke(t *testing.T) {
	cfg := loadTestConfig()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	requireServiceReady(t, ctx, client, cfg.TaifaIDBaseURL, "TaifaID", cfg.Force)
	requireServiceReady(t, ctx, client, cfg.TaifaExchangeBaseURL, "TaifaExchange", cfg.Force)

	token := loginToTaifaID(t, ctx, client, cfg)

	t.Run("clinician can create care claim", func(t *testing.T) {
		response := authorize(t, ctx, client, cfg, token, map[string]string{
			"organization_id": cfg.OrganizationID,
			"target_system":   "TAIFA_CARE",
			"route":           "/api/v1/claims",
			"method":          "POST",
			"operation":       "care.claim.create",
		})

		if response.Data.Decision != "ALLOW" {
			t.Fatalf("expected ALLOW, got %q", response.Data.Decision)
		}

		if response.Data.MatchedPolicyID != "POL-CARE-CLAIM-CREATE-CLINICIAN" {
			t.Fatalf("expected matched clinician policy, got %q", response.Data.MatchedPolicyID)
		}

		if response.Data.DecisionID == "" {
			t.Fatalf("expected decision_id to be populated")
		}

		if response.Data.ActorContext.PersonID == "" {
			t.Fatalf("expected actor_context.person_id to be populated")
		}

		if response.Data.ActorContext.OrganizationID != cfg.OrganizationID {
			t.Fatalf("expected organization_id %q, got %q", cfg.OrganizationID, response.Data.ActorContext.OrganizationID)
		}

		if !contains(response.Data.ActorContext.Roles, "PROVIDER_CLINICIAN") {
			t.Fatalf("expected PROVIDER_CLINICIAN role, got %#v", response.Data.ActorContext.Roles)
		}

		if !response.Data.Obligations.PropagateCorrelationID {
			t.Fatalf("expected propagate_correlation_id obligation")
		}

		if !response.Data.Obligations.RequireAudit {
			t.Fatalf("expected require_audit obligation")
		}
	})

	t.Run("clinician cannot create payment instruction", func(t *testing.T) {
		envelope, statusCode := authorizeExpectError(t, ctx, client, cfg, token, map[string]string{
			"organization_id": cfg.OrganizationID,
			"target_system":   "TAIFA_PAY",
			"route":           "/api/v1/payment-instructions",
			"method":          "POST",
			"operation":       "pay.instruction.create",
		})

		if statusCode != http.StatusForbidden {
			t.Fatalf("expected HTTP 403, got %d", statusCode)
		}

		if envelope.Error.Code != "FORBIDDEN" {
			t.Fatalf("expected FORBIDDEN error code, got %q", envelope.Error.Code)
		}

		if envelope.CorrelationID == "" {
			t.Fatalf("expected correlation_id to be populated")
		}
	})
}

func loadTestConfig() testConfig {
	return testConfig{
		TaifaIDBaseURL:       envString("TAIFA_EXCHANGE_TEST_TAIFA_ID_BASE_URL", defaultTaifaIDBaseURL),
		TaifaExchangeBaseURL: envString("TAIFA_EXCHANGE_TEST_BASE_URL", defaultTaifaExchangeBaseURL),
		Username:             envString("TAIFA_EXCHANGE_TEST_USERNAME", defaultUsername),
		Password:             envString("TAIFA_EXCHANGE_TEST_PASSWORD", defaultPassword),
		OrganizationID:       envString("TAIFA_EXCHANGE_TEST_ORGANIZATION_ID", defaultOrganizationID),
		Force:                envBool("TAIFA_EXCHANGE_RUN_INTEGRATION_TESTS", false),
	}
}

func requireServiceReady(
	t *testing.T,
	ctx context.Context,
	client *http.Client,
	baseURL string,
	name string,
	force bool,
) {
	t.Helper()

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, trimBaseURL(baseURL)+"/readyz", nil)
	if err != nil {
		t.Fatalf("create %s readiness request: %v", name, err)
	}

	response, err := client.Do(request)
	if err != nil {
		if force {
			t.Fatalf("%s is not reachable: %v", name, err)
		}

		t.Skipf("%s is not reachable; set TAIFA_EXCHANGE_RUN_INTEGRATION_TESTS=true to force failure: %v", name, err)
	}
	defer response.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(response.Body, 1<<20))

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		if force {
			t.Fatalf("%s readiness failed: status=%d body=%s", name, response.StatusCode, strings.TrimSpace(string(body)))
		}

		t.Skipf("%s readiness failed; set TAIFA_EXCHANGE_RUN_INTEGRATION_TESTS=true to force failure: status=%d body=%s", name, response.StatusCode, strings.TrimSpace(string(body)))
	}
}

func loginToTaifaID(
	t *testing.T,
	ctx context.Context,
	client *http.Client,
	cfg testConfig,
) string {
	t.Helper()

	payload := map[string]string{
		"username": cfg.Username,
		"password": cfg.Password,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal login payload: %v", err)
	}

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		trimBaseURL(cfg.TaifaIDBaseURL)+"/api/v1/auth/login",
		bytes.NewReader(body),
	)
	if err != nil {
		t.Fatalf("create login request: %v", err)
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")

	response, err := client.Do(request)
	if err != nil {
		t.Fatalf("call TaifaID login: %v", err)
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	if err != nil {
		t.Fatalf("read login response: %v", err)
	}

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		t.Fatalf("login failed: status=%d body=%s", response.StatusCode, strings.TrimSpace(string(responseBody)))
	}

	var envelope loginEnvelope
	if err := json.Unmarshal(responseBody, &envelope); err != nil {
		t.Fatalf("decode login envelope: %v body=%s", err, strings.TrimSpace(string(responseBody)))
	}

	token := extractToken(envelope.Data)
	if token == "" {
		t.Fatalf("login response did not contain token/access_token/jwt: body=%s", strings.TrimSpace(string(responseBody)))
	}

	return token
}

func authorize(
	t *testing.T,
	ctx context.Context,
	client *http.Client,
	cfg testConfig,
	token string,
	payload map[string]string,
) authorizeEnvelope {
	t.Helper()

	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal authorize payload: %v", err)
	}

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		trimBaseURL(cfg.TaifaExchangeBaseURL)+"/api/v1/exchange/authorize",
		bytes.NewReader(body),
	)
	if err != nil {
		t.Fatalf("create authorize request: %v", err)
	}

	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")

	response, err := client.Do(request)
	if err != nil {
		t.Fatalf("call TaifaExchange authorize: %v", err)
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	if err != nil {
		t.Fatalf("read authorize response: %v", err)
	}

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		t.Fatalf("authorize failed: status=%d body=%s", response.StatusCode, strings.TrimSpace(string(responseBody)))
	}

	var envelope authorizeEnvelope
	if err := json.Unmarshal(responseBody, &envelope); err != nil {
		t.Fatalf("decode authorize response: %v body=%s", err, strings.TrimSpace(string(responseBody)))
	}

	return envelope
}

func authorizeExpectError(
	t *testing.T,
	ctx context.Context,
	client *http.Client,
	cfg testConfig,
	token string,
	payload map[string]string,
) (errorEnvelope, int) {
	t.Helper()

	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal authorize payload: %v", err)
	}

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		trimBaseURL(cfg.TaifaExchangeBaseURL)+"/api/v1/exchange/authorize",
		bytes.NewReader(body),
	)
	if err != nil {
		t.Fatalf("create authorize request: %v", err)
	}

	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")

	response, err := client.Do(request)
	if err != nil {
		t.Fatalf("call TaifaExchange authorize: %v", err)
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	if err != nil {
		t.Fatalf("read authorize error response: %v", err)
	}

	if response.StatusCode >= 200 && response.StatusCode < 300 {
		t.Fatalf("expected authorize error, got status=%d body=%s", response.StatusCode, strings.TrimSpace(string(responseBody)))
	}

	var envelope errorEnvelope
	if err := json.Unmarshal(responseBody, &envelope); err != nil {
		t.Fatalf("decode authorize error response: %v body=%s", err, strings.TrimSpace(string(responseBody)))
	}

	return envelope, response.StatusCode
}

func extractToken(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}

	var data map[string]any
	if err := json.Unmarshal(raw, &data); err != nil {
		return ""
	}

	for _, key := range []string{"token", "access_token", "jwt"} {
		if value, ok := data[key].(string); ok && strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}

	return findTokenRecursive(data)
}

func findTokenRecursive(value any) string {
	switch typed := value.(type) {
	case map[string]any:
		for _, key := range []string{"token", "access_token", "jwt"} {
			if value, ok := typed[key].(string); ok && strings.TrimSpace(value) != "" {
				return strings.TrimSpace(value)
			}
		}

		for _, nested := range typed {
			if token := findTokenRecursive(nested); token != "" {
				return token
			}
		}

	case []any:
		for _, nested := range typed {
			if token := findTokenRecursive(nested); token != "" {
				return token
			}
		}
	}

	return ""
}

func contains(values []string, expected string) bool {
	expected = strings.ToUpper(strings.TrimSpace(expected))

	for _, value := range values {
		if strings.ToUpper(strings.TrimSpace(value)) == expected {
			return true
		}
	}

	return false
}

func trimBaseURL(value string) string {
	return strings.TrimRight(strings.TrimSpace(value), "/")
}

func envString(key string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	return value
}

func envBool(key string, fallback bool) bool {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	switch strings.ToLower(value) {
	case "1", "true", "yes", "y", "on":
		return true
	case "0", "false", "no", "n", "off":
		return false
	default:
		return fallback
	}
}
