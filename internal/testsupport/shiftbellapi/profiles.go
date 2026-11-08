package shiftbellapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type Profile struct {
	Id      int       `json:"id"`
	Name    string    `json:"name"`
	Links   Relations `json:"_links"`
	Actions Relations `json:"_actions"`
}

type ProfileCollection struct {
	Items   []Profile `json:"items"`
	Links   Relations `json:"_links"`
	Actions Relations `json:"_actions"`
}

func (c *APIClient) DeleteProfile(ctx context.Context, href string) error {
	request, err := c.newRequest(ctx, http.MethodDelete, href, nil)
	if err != nil {
		return err
	}
	_, _, err = c.do(request, http.StatusNoContent)
	return err
}

type ProfileSetup struct {
	Required bool      `json:"required"`
	Links    Relations `json:"_links"`
	Actions  Relations `json:"_actions"`
}

func (c *APIClient) GetProfiles(
	ctx context.Context,
	href string,
) (*ProfileCollection, error) {
	request, err := c.newRequest(ctx, http.MethodGet, href, nil)
	if err != nil {
		return nil, err
	}
	_, body, err := c.do(request, http.StatusOK)
	if err != nil {
		return nil, err
	}
	var result ProfileCollection
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("decode profile collection: %w", err)
	}
	return &result, nil
}

func (c *APIClient) GetProfileSetup(
	ctx context.Context,
	href string,
) (*ProfileSetup, error) {
	request, err := c.newRequest(ctx, http.MethodGet, href, nil)
	if err != nil {
		return nil, err
	}
	_, body, err := c.do(request, http.StatusOK)
	if err != nil {
		return nil, err
	}
	var result ProfileSetup
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("decode profile setup: %w", err)
	}
	return &result, nil
}

func (c *APIClient) CreateProfile(
	ctx context.Context,
	href, name string,
) (*Profile, error) {
	body, err := json.Marshal(struct {
		Name string `json:"name"`
	}{Name: name})
	if err != nil {
		return nil, fmt.Errorf("encode profile name: %w", err)
	}
	request, err := c.newRequest(ctx, http.MethodPost, href, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json")
	_, body, err = c.do(request, http.StatusCreated)
	if err != nil {
		return nil, err
	}
	var result Profile
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("decode created profile: %w", err)
	}
	return &result, nil
}
