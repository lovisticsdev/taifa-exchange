package taifaid

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string, timeout time.Duration) *Client {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")

	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *Client) IsConfigured() bool {
	return c != nil && c.baseURL != ""
}

func (c *Client) Ready(ctx context.Context, correlationID string) (*ReadyResponse, error) {
	if !c.IsConfigured() {
		return nil, fmt.Errorf("taifa-id client is not configured")
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/readyz", nil)
	if err != nil {
		return nil, fmt.Errorf("create taifa-id readiness request: %w", err)
	}

	if correlationID != "" {
		request.Header.Set("X-Correlation-ID", correlationID)
	}

	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("call taifa-id readiness: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, c.decodeHTTPError(response, "taifa-id readiness failed")
	}

	var ready ReadyResponse
	if err := json.NewDecoder(response.Body).Decode(&ready); err != nil {
		return nil, fmt.Errorf("decode taifa-id readiness response: %w", err)
	}

	if ready.Status != "ok" {
		return nil, fmt.Errorf("taifa-id readiness status is %q", ready.Status)
	}

	return &ready, nil
}

func (c *Client) ResolveActorContext(
	ctx context.Context,
	token string,
	organizationID string,
	correlationID string,
) (*ActorContextResponse, error) {
	if !c.IsConfigured() {
		return nil, fmt.Errorf("taifa-id client is not configured")
	}

	token = strings.TrimSpace(token)
	organizationID = strings.TrimSpace(organizationID)

	if token == "" {
		return nil, fmt.Errorf("token is required")
	}

	if organizationID == "" {
		return nil, fmt.Errorf("organization_id is required")
	}

	requestBody := ResolveActorContextRequest{
		Token:          token,
		OrganizationID: organizationID,
	}

	var responseEnvelope ActorContextResponse
	if err := c.postJSON(
		ctx,
		"/api/v1/actor-context/resolve",
		correlationID,
		requestBody,
		&responseEnvelope,
	); err != nil {
		return nil, err
	}

	return &responseEnvelope, nil
}

func (c *Client) ListOrganizationCapabilities(
	ctx context.Context,
	organizationID string,
	correlationID string,
) (*CapabilitiesResponse, error) {
	if !c.IsConfigured() {
		return nil, fmt.Errorf("taifa-id client is not configured")
	}

	organizationID = strings.TrimSpace(organizationID)
	if organizationID == "" {
		return nil, fmt.Errorf("organization_id is required")
	}

	path := "/api/v1/organizations/" + url.PathEscape(organizationID) + "/capabilities"

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("create taifa-id capabilities request: %w", err)
	}

	if correlationID != "" {
		request.Header.Set("X-Correlation-ID", correlationID)
	}

	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("call taifa-id capabilities: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, c.decodeHTTPError(response, "taifa-id capabilities lookup failed")
	}

	var responseEnvelope CapabilitiesResponse
	if err := json.NewDecoder(response.Body).Decode(&responseEnvelope); err != nil {
		return nil, fmt.Errorf("decode taifa-id capabilities response: %w", err)
	}

	return &responseEnvelope, nil
}

func (c *Client) postJSON(
	ctx context.Context,
	path string,
	correlationID string,
	requestPayload any,
	responsePayload any,
) error {
	body, err := json.Marshal(requestPayload)
	if err != nil {
		return fmt.Errorf("marshal taifa-id request: %w", err)
	}

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.baseURL+path,
		bytes.NewReader(body),
	)
	if err != nil {
		return fmt.Errorf("create taifa-id request: %w", err)
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")

	if correlationID != "" {
		request.Header.Set("X-Correlation-ID", correlationID)
	}

	response, err := c.httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("call taifa-id: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return c.decodeHTTPError(response, "taifa-id request failed")
	}

	if err := json.NewDecoder(response.Body).Decode(responsePayload); err != nil {
		return fmt.Errorf("decode taifa-id response: %w", err)
	}

	return nil
}

func (c *Client) decodeHTTPError(response *http.Response, fallback string) error {
	body, readErr := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	if readErr != nil {
		return fmt.Errorf("%s: status=%d", fallback, response.StatusCode)
	}

	var errorEnvelope ErrorEnvelope
	if err := json.Unmarshal(body, &errorEnvelope); err == nil {
		if errorEnvelope.Error.Message != "" {
			return fmt.Errorf(
				"%s: status=%d code=%s message=%s",
				fallback,
				response.StatusCode,
				errorEnvelope.Error.Code,
				errorEnvelope.Error.Message,
			)
		}
	}

	bodyText := strings.TrimSpace(string(body))
	if bodyText == "" {
		return fmt.Errorf("%s: status=%d", fallback, response.StatusCode)
	}

	return fmt.Errorf("%s: status=%d body=%s", fallback, response.StatusCode, bodyText)
}
