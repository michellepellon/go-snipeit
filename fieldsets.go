package snipeit

import (
	"context"
	"fmt"
	"net/http"
)

// FieldsetsService handles communication with the custom fieldset-related endpoints
// of the Snipe-IT API.
//
// Snipe-IT API docs: https://snipe-it.readme.io/reference/fieldsets
type FieldsetsService struct {
	client *Client
}

// FieldsetResponse represents the API response for a single fieldset.
type FieldsetResponse struct {
	Response
	Payload Fieldset `json:"payload"`
}

// FieldsetsResponse represents the API response for multiple fieldsets.
type FieldsetsResponse struct {
	Response
	Rows []Fieldset `json:"rows"`
}

// List returns a list of custom fieldsets with pagination options.
func (s *FieldsetsService) List(opts *ListOptions) (*FieldsetsResponse, *http.Response, error) {
	return s.ListContext(context.Background(), opts)
}

// ListContext returns a list of custom fieldsets with the provided context and pagination options.
func (s *FieldsetsService) ListContext(ctx context.Context, opts *ListOptions) (*FieldsetsResponse, *http.Response, error) {
	u := "api/v1/fieldsets"
	if opts != nil {
		var err error
		u, err = s.client.AddOptions(u, opts)
		if err != nil {
			return nil, nil, err
		}
	}

	req, err := s.client.newRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, nil, err
	}

	var fieldsets FieldsetsResponse
	resp, err := s.client.Do(req, &fieldsets)
	if err != nil {
		return nil, resp, err
	}

	return &fieldsets, resp, nil
}

// Get fetches a single custom fieldset by its ID.
func (s *FieldsetsService) Get(id int) (*FieldsetResponse, *http.Response, error) {
	return s.GetContext(context.Background(), id)
}

// GetContext fetches a single custom fieldset by its ID with the provided context.
func (s *FieldsetsService) GetContext(ctx context.Context, id int) (*FieldsetResponse, *http.Response, error) {
	u := fmt.Sprintf("api/v1/fieldsets/%d", id)
	req, err := s.client.newRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, nil, err
	}

	var fieldset FieldsetResponse
	resp, err := s.client.Do(req, &fieldset)
	if err != nil {
		return nil, resp, err
	}

	return &fieldset, resp, nil
}

// Create creates a new custom fieldset in Snipe-IT.
func (s *FieldsetsService) Create(fieldset Fieldset) (*FieldsetResponse, *http.Response, error) {
	return s.CreateContext(context.Background(), fieldset)
}

// CreateContext creates a new custom fieldset in Snipe-IT with the provided context.
func (s *FieldsetsService) CreateContext(ctx context.Context, fieldset Fieldset) (*FieldsetResponse, *http.Response, error) {
	req, err := s.client.newRequestWithContext(ctx, http.MethodPost, "api/v1/fieldsets", fieldset)
	if err != nil {
		return nil, nil, err
	}

	var response FieldsetResponse
	resp, err := s.client.Do(req, &response)
	if err != nil {
		return nil, resp, err
	}

	return &response, resp, nil
}

// Update updates an existing custom fieldset in Snipe-IT.
func (s *FieldsetsService) Update(id int, fieldset Fieldset) (*FieldsetResponse, *http.Response, error) {
	return s.UpdateContext(context.Background(), id, fieldset)
}

// UpdateContext updates an existing custom fieldset in Snipe-IT with the provided context.
func (s *FieldsetsService) UpdateContext(ctx context.Context, id int, fieldset Fieldset) (*FieldsetResponse, *http.Response, error) {
	u := fmt.Sprintf("api/v1/fieldsets/%d", id)
	req, err := s.client.newRequestWithContext(ctx, http.MethodPut, u, fieldset)
	if err != nil {
		return nil, nil, err
	}

	var response FieldsetResponse
	resp, err := s.client.Do(req, &response)
	if err != nil {
		return nil, resp, err
	}

	return &response, resp, nil
}

// Delete deletes a custom fieldset from Snipe-IT.
func (s *FieldsetsService) Delete(id int) (*http.Response, error) {
	return s.DeleteContext(context.Background(), id)
}

// DeleteContext deletes a custom fieldset from Snipe-IT with the provided context.
func (s *FieldsetsService) DeleteContext(ctx context.Context, id int) (*http.Response, error) {
	u := fmt.Sprintf("api/v1/fieldsets/%d", id)
	req, err := s.client.newRequestWithContext(ctx, http.MethodDelete, u, nil)
	if err != nil {
		return nil, err
	}

	return s.client.Do(req, nil)
}
