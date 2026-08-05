package shiftbellapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	MediaType                 = "application/vnd.shiftbell+json"
	RelationSelf              = "self"
	RelationCollection        = "collection"
	RelationChores            = "chores"
	RelationChoreTemplates    = "chore_templates"
	RelationNext              = "next"
	RelationPrevious          = "previous"
	ActionCreateChoreTemplate = "create"
	ActionCreateChore         = "create"
	ActionEditChore           = "edit"
	ActionCompleteChore       = "complete"
	ActionCorrectCompletion   = "correct_completion"
	ActionDeleteChore         = "delete"
	ActionEditChoreTemplate   = "edit"
	ActionDeactivateTemplate  = "deactivate"
)

type APIClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewAPIClient(baseURL string) *APIClient {
	return &APIClient{
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}
}

func decodeErrorResponse(responseBody []byte) (*ErrorResponse, error) {
	var response ErrorResponse
	if err := json.Unmarshal(responseBody, &response); err != nil {
		return nil, fmt.Errorf("decode error response: %w", err)
	}
	return &response, nil
}

func (c *APIClient) newRequest(
	ctx context.Context,
	method string,
	href string,
	body io.Reader,
) (*http.Request, error) {
	baseURL, err := url.Parse(c.baseURL + "/")
	if err != nil {
		return nil, fmt.Errorf("parse base URL: %w", err)
	}
	reference, err := url.Parse(href)
	if err != nil {
		return nil, fmt.Errorf("parse reference: %w", err)
	}
	request, err := http.NewRequestWithContext(
		ctx,
		method,
		baseURL.ResolveReference(reference).String(),
		body,
	)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Accept", MediaType)
	return request, nil
}

func (c *APIClient) do(
	request *http.Request,
	expectedStatus int,
) (http.Header, []byte, error) {
	statusCode, header, responseBody, err := c.doResponse(request)
	if err != nil {
		return nil, nil, err
	}
	if statusCode != expectedStatus {
		return nil, nil, fmt.Errorf("unexpected status %d: %s", statusCode, responseBody)
	}
	return header, responseBody, nil
}

func (c *APIClient) doResponse(request *http.Request) (int, http.Header, []byte, error) {
	response, err := c.httpClient.Do(request)
	if err != nil {
		return 0, nil, nil, fmt.Errorf("send request: %w", err)
	}
	responseBody, readErr := io.ReadAll(response.Body)
	closeErr := response.Body.Close()
	if readErr != nil {
		return 0, nil, nil, fmt.Errorf("read response: %w", readErr)
	}
	if closeErr != nil {
		return 0, nil, nil, fmt.Errorf("close response: %w", closeErr)
	}
	if response.StatusCode == http.StatusNoContent {
		return response.StatusCode, response.Header, responseBody, nil
	}
	mediaType, _, err := mime.ParseMediaType(response.Header.Get("Content-Type"))
	if err != nil {
		return 0, nil, nil, fmt.Errorf("parse response content type: %w", err)
	}
	if mediaType != MediaType {
		return 0, nil, nil, fmt.Errorf("unexpected response content type %q", mediaType)
	}
	return response.StatusCode, response.Header, responseBody, nil
}
