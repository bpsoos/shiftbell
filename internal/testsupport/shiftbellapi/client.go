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
	RelationExisting          = "existing"
	RelationNext              = "next"
	RelationPrevious          = "previous"
	ActionCreateChoreTemplate = "create"
	ActionEditChoreTemplate   = "edit"
	ActionDeactivateTemplate  = "deactivate"
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

type BrowseChoreTemplatesParams struct {
	Href   string
	Search string
	State  string
	Limit  int
}

type GetChoreTemplatesResult struct {
	Collection ChoreTemplateCollection
}

type CreateChoreTemplateParams struct {
	Action      Action `json:"-"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type CreateChoreTemplateResponse struct {
	ChoreTemplate ChoreTemplate
	Actions       Actions
	Location      string
}

type CreateChoreTemplateResult struct {
	StatusCode      int
	SuccessResponse *CreateChoreTemplateResponse
	ErrorResponse   *ErrorResponse
}

type GetChoreTemplateParams struct {
	Link Link
}

type GetChoreTemplateResponse struct {
	ChoreTemplate ChoreTemplate
	Actions       Actions
}

type GetChoreTemplateResult struct {
	StatusCode      int
	SuccessResponse *GetChoreTemplateResponse
	ErrorResponse   *ErrorResponse
}

type EditChoreTemplateParams struct {
	Action      Action `json:"-"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type EditChoreTemplateResponse struct {
	ChoreTemplate ChoreTemplate
	Actions       Actions
}

type EditChoreTemplateResult struct {
	StatusCode      int
	SuccessResponse *EditChoreTemplateResponse
	ErrorResponse   *ErrorResponse
}

type DeactivateChoreTemplateParams struct {
	Action Action
}

type DeactivateChoreTemplateResult struct {
	ChoreTemplate ChoreTemplate
	Actions       Actions
}

type ErrorResponse struct {
	Error   string  `json:"error"`
	Links   Links   `json:"_links"`
	Actions Actions `json:"_actions"`
}

type choreTemplateRepresentation struct {
	ChoreTemplate
	Actions Actions `json:"_actions"`
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

func (c *APIClient) BrowseChoreTemplates(ctx context.Context, params BrowseChoreTemplatesParams) (*GetChoreTemplatesResult, error) {
	reference, err := url.Parse(params.Href)
	if err != nil {
		return nil, fmt.Errorf("parse chore template collection reference: %w", err)
	}
	query := reference.Query()
	if params.Search != "" {
		query.Set("search", params.Search)
	}
	if params.State != "" {
		query.Set("state", params.State)
	}
	if params.Limit > 0 {
		query.Set("limit", fmt.Sprintf("%d", params.Limit))
	}
	reference.RawQuery = query.Encode()
	return c.GetChoreTemplates(ctx, reference.String())
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

	statusCode, header, responseBody, err := c.doResponse(request)
	if err != nil {
		return nil, err
	}
	result := &CreateChoreTemplateResult{StatusCode: statusCode}
	if statusCode != http.StatusCreated {
		result.ErrorResponse, err = decodeErrorResponse(responseBody)
		if err != nil {
			return nil, err
		}
		return result, nil
	}
	var representation choreTemplateRepresentation
	if err := json.Unmarshal(responseBody, &representation); err != nil {
		return nil, fmt.Errorf("decode create chore template response: %w", err)
	}

	result.SuccessResponse = &CreateChoreTemplateResponse{
		ChoreTemplate: representation.ChoreTemplate,
		Actions:       representation.Actions,
		Location:      header.Get("Location"),
	}
	return result, nil
}

func (c *APIClient) GetChoreTemplate(ctx context.Context, params GetChoreTemplateParams) (*GetChoreTemplateResult, error) {
	request, err := c.newRequest(ctx, http.MethodGet, params.Link.Href, nil)
	if err != nil {
		return nil, fmt.Errorf("create get chore template request: %w", err)
	}

	statusCode, _, responseBody, err := c.doResponse(request)
	if err != nil {
		return nil, err
	}
	result := &GetChoreTemplateResult{StatusCode: statusCode}
	if statusCode != http.StatusOK {
		result.ErrorResponse, err = decodeErrorResponse(responseBody)
		if err != nil {
			return nil, err
		}
		return result, nil
	}
	var representation choreTemplateRepresentation
	if err := json.Unmarshal(responseBody, &representation); err != nil {
		return nil, fmt.Errorf("decode get chore template response: %w", err)
	}

	result.SuccessResponse = &GetChoreTemplateResponse{
		ChoreTemplate: representation.ChoreTemplate,
		Actions:       representation.Actions,
	}
	return result, nil
}

func (c *APIClient) EditChoreTemplate(ctx context.Context, params EditChoreTemplateParams) (*EditChoreTemplateResult, error) {
	requestBody, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("encode edit chore template request: %w", err)
	}
	request, err := c.newRequest(ctx, params.Action.Method, params.Action.Href, bytes.NewReader(requestBody))
	if err != nil {
		return nil, fmt.Errorf("create edit chore template request: %w", err)
	}
	request.Header.Set("Content-Type", params.Action.ContentType)

	statusCode, _, responseBody, err := c.doResponse(request)
	if err != nil {
		return nil, err
	}
	result := &EditChoreTemplateResult{StatusCode: statusCode}
	if statusCode != http.StatusOK {
		result.ErrorResponse, err = decodeErrorResponse(responseBody)
		if err != nil {
			return nil, err
		}
		return result, nil
	}
	var representation choreTemplateRepresentation
	if err := json.Unmarshal(responseBody, &representation); err != nil {
		return nil, fmt.Errorf("decode edit chore template response: %w", err)
	}
	result.SuccessResponse = &EditChoreTemplateResponse{
		ChoreTemplate: representation.ChoreTemplate,
		Actions:       representation.Actions,
	}
	return result, nil
}

func (c *APIClient) DeactivateChoreTemplate(ctx context.Context, params DeactivateChoreTemplateParams) (*DeactivateChoreTemplateResult, error) {
	request, err := c.newRequest(ctx, params.Action.Method, params.Action.Href, nil)
	if err != nil {
		return nil, fmt.Errorf("create deactivate chore template request: %w", err)
	}
	if params.Action.ContentType != "" {
		request.Header.Set("Content-Type", params.Action.ContentType)
	}

	_, responseBody, err := c.do(request, http.StatusOK)
	if err != nil {
		return nil, err
	}
	var representation choreTemplateRepresentation
	if err := json.Unmarshal(responseBody, &representation); err != nil {
		return nil, fmt.Errorf("decode deactivate chore template response: %w", err)
	}
	return &DeactivateChoreTemplateResult{
		ChoreTemplate: representation.ChoreTemplate,
		Actions:       representation.Actions,
	}, nil
}

func decodeErrorResponse(responseBody []byte) (*ErrorResponse, error) {
	var response ErrorResponse
	if err := json.Unmarshal(responseBody, &response); err != nil {
		return nil, fmt.Errorf("decode error response: %w", err)
	}
	return &response, nil
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
	mediaType, _, err := mime.ParseMediaType(response.Header.Get("Content-Type"))
	if err != nil {
		return 0, nil, nil, fmt.Errorf("parse response content type: %w", err)
	}
	if mediaType != MediaType {
		return 0, nil, nil, fmt.Errorf("unexpected response content type %q", mediaType)
	}

	return response.StatusCode, response.Header, responseBody, nil
}
