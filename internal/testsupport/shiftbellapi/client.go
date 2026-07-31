package shiftbellapi

import (
	"bytes"
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
	RelationChoreTemplates    = "chore_templates"
	ActionCreateChoreTemplate = "create"
)

type APIClient struct {
	baseURL    string
	httpClient *http.Client
}

type Link struct {
	Href string `json:"href"`
}

type Links map[string]Link

type ActionField struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	Required  bool   `json:"required"`
	MaxLength int    `json:"max_length"`
}

type Action struct {
	Href        string        `json:"href"`
	Method      string        `json:"method"`
	ContentType string        `json:"content_type"`
	Fields      []ActionField `json:"fields"`
}

type Actions map[string]Action

type Home struct {
	Links Links `json:"_links"`
}

type GetHomeResult struct {
	Home Home
}

type ChoreTemplate struct {
	Id            int     `json:"id"`
	Name          string  `json:"name"`
	Description   string  `json:"description"`
	DeactivatedAt *string `json:"deactivated_at"`
	Links         Links   `json:"_links"`
}

type ChoreTemplateCollection struct {
	Items   []ChoreTemplate `json:"items"`
	More    bool            `json:"more"`
	Links   Links           `json:"_links"`
	Actions Actions         `json:"_actions"`
}

type GetChoreTemplatesResult struct {
	Collection ChoreTemplateCollection
}

type CreateChoreTemplateParams struct {
	Action      Action `json:"-"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type CreateChoreTemplateResult struct {
	ChoreTemplate ChoreTemplate
	Location      string
}

type GetChoreTemplateParams struct {
	Link Link
}

type GetChoreTemplateResult struct {
	ChoreTemplate ChoreTemplate
}

func NewAPIClient(baseURL string) *APIClient {
	return &APIClient{
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}
}

func (c *APIClient) GetHome(ctx context.Context) (*GetHomeResult, error) {
	request, err := c.newRequest(ctx, http.MethodGet, "/", nil)
	if err != nil {
		return nil, fmt.Errorf("create get home request: %w", err)
	}

	_, responseBody, err := c.do(request, http.StatusOK)
	if err != nil {
		return nil, err
	}
	var home Home
	if err := json.Unmarshal(responseBody, &home); err != nil {
		return nil, fmt.Errorf("decode home response: %w", err)
	}

	return &GetHomeResult{Home: home}, nil
}

func (c *APIClient) GetChoreTemplates(ctx context.Context, href string) (*GetChoreTemplatesResult, error) {
	request, err := c.newRequest(ctx, http.MethodGet, href, nil)
	if err != nil {
		return nil, fmt.Errorf("create get chore templates request: %w", err)
	}

	_, responseBody, err := c.do(request, http.StatusOK)
	if err != nil {
		return nil, err
	}
	var collection ChoreTemplateCollection
	if err := json.Unmarshal(responseBody, &collection); err != nil {
		return nil, fmt.Errorf("decode chore template collection response: %w", err)
	}

	return &GetChoreTemplatesResult{Collection: collection}, nil
}

func (c *APIClient) CreateChoreTemplate(ctx context.Context, params CreateChoreTemplateParams) (*CreateChoreTemplateResult, error) {
	requestBody, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("encode create chore template request: %w", err)
	}
	request, err := c.newRequest(ctx, params.Action.Method, params.Action.Href, bytes.NewReader(requestBody))
	if err != nil {
		return nil, fmt.Errorf("create chore template request: %w", err)
	}
	request.Header.Set("Content-Type", params.Action.ContentType)

	header, responseBody, err := c.do(request, http.StatusCreated)
	if err != nil {
		return nil, err
	}
	var choreTemplate ChoreTemplate
	if err := json.Unmarshal(responseBody, &choreTemplate); err != nil {
		return nil, fmt.Errorf("decode create chore template response: %w", err)
	}

	return &CreateChoreTemplateResult{
		ChoreTemplate: choreTemplate,
		Location:      header.Get("Location"),
	}, nil
}

func (c *APIClient) GetChoreTemplate(ctx context.Context, params GetChoreTemplateParams) (*GetChoreTemplateResult, error) {
	request, err := c.newRequest(ctx, http.MethodGet, params.Link.Href, nil)
	if err != nil {
		return nil, fmt.Errorf("create get chore template request: %w", err)
	}

	_, responseBody, err := c.do(request, http.StatusOK)
	if err != nil {
		return nil, err
	}
	var choreTemplate ChoreTemplate
	if err := json.Unmarshal(responseBody, &choreTemplate); err != nil {
		return nil, fmt.Errorf("decode get chore template response: %w", err)
	}

	return &GetChoreTemplateResult{ChoreTemplate: choreTemplate}, nil
}

func (c *APIClient) newRequest(ctx context.Context, method string, href string, body io.Reader) (*http.Request, error) {
	baseURL, err := url.Parse(c.baseURL + "/")
	if err != nil {
		return nil, fmt.Errorf("parse base URL: %w", err)
	}
	reference, err := url.Parse(href)
	if err != nil {
		return nil, fmt.Errorf("parse reference: %w", err)
	}
	request, err := http.NewRequestWithContext(ctx, method, baseURL.ResolveReference(reference).String(), body)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Accept", MediaType)
	return request, nil
}

func (c *APIClient) do(request *http.Request, expectedStatus int) (http.Header, []byte, error) {
	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, nil, fmt.Errorf("send request: %w", err)
	}
	responseBody, readErr := io.ReadAll(response.Body)
	closeErr := response.Body.Close()
	if readErr != nil {
		return nil, nil, fmt.Errorf("read response: %w", readErr)
	}
	if closeErr != nil {
		return nil, nil, fmt.Errorf("close response: %w", closeErr)
	}
	if response.StatusCode != expectedStatus {
		return nil, nil, fmt.Errorf("unexpected status %d: %s", response.StatusCode, responseBody)
	}
	mediaType, _, err := mime.ParseMediaType(response.Header.Get("Content-Type"))
	if err != nil {
		return nil, nil, fmt.Errorf("parse response content type: %w", err)
	}
	if mediaType != MediaType {
		return nil, nil, fmt.Errorf("unexpected response content type %q", mediaType)
	}

	return response.Header, responseBody, nil
}
