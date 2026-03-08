package snipeit

import (
	"context"
	"fmt"
	"net/http"
)

// FieldsService handles communication with the custom field-related endpoints
// of the Snipe-IT API.
//
// Snipe-IT API docs: https://snipe-it.readme.io/reference/fields
type FieldsService struct {
	client *Client
}

// FieldResponse represents the API response for a single field.
type FieldResponse struct {
	Response
	Payload Field `json:"payload"`
}

// FieldsResponse represents the API response for multiple fields.
type FieldsResponse struct {
	Response
	Rows []Field `json:"rows"`
}

// List returns a list of custom fields with pagination options.
func (s *FieldsService) List(opts *ListOptions) (*FieldsResponse, *http.Response, error) {
	return s.ListContext(context.Background(), opts)
}

// ListContext returns a list of custom fields with the provided context and pagination options.
func (s *FieldsService) ListContext(ctx context.Context, opts *ListOptions) (*FieldsResponse, *http.Response, error) {
	u := "api/v1/fields"
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

	var fields FieldsResponse
	resp, err := s.client.Do(req, &fields)
	if err != nil {
		return nil, resp, err
	}

	return &fields, resp, nil
}

// Get fetches a single custom field by its ID.
func (s *FieldsService) Get(id int) (*FieldResponse, *http.Response, error) {
	return s.GetContext(context.Background(), id)
}

// GetContext fetches a single custom field by its ID with the provided context.
func (s *FieldsService) GetContext(ctx context.Context, id int) (*FieldResponse, *http.Response, error) {
	u := fmt.Sprintf("api/v1/fields/%d", id)
	req, err := s.client.newRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, nil, err
	}

	var field FieldResponse
	resp, err := s.client.Do(req, &field)
	if err != nil {
		return nil, resp, err
	}

	return &field, resp, nil
}

// Create creates a new custom field in Snipe-IT.
func (s *FieldsService) Create(field Field) (*FieldResponse, *http.Response, error) {
	return s.CreateContext(context.Background(), field)
}

// CreateContext creates a new custom field in Snipe-IT with the provided context.
func (s *FieldsService) CreateContext(ctx context.Context, field Field) (*FieldResponse, *http.Response, error) {
	req, err := s.client.newRequestWithContext(ctx, http.MethodPost, "api/v1/fields", field)
	if err != nil {
		return nil, nil, err
	}

	var response FieldResponse
	resp, err := s.client.Do(req, &response)
	if err != nil {
		return nil, resp, err
	}

	return &response, resp, nil
}

// Update updates an existing custom field in Snipe-IT.
func (s *FieldsService) Update(id int, field Field) (*FieldResponse, *http.Response, error) {
	return s.UpdateContext(context.Background(), id, field)
}

// UpdateContext updates an existing custom field in Snipe-IT with the provided context.
func (s *FieldsService) UpdateContext(ctx context.Context, id int, field Field) (*FieldResponse, *http.Response, error) {
	u := fmt.Sprintf("api/v1/fields/%d", id)
	req, err := s.client.newRequestWithContext(ctx, http.MethodPut, u, field)
	if err != nil {
		return nil, nil, err
	}

	var response FieldResponse
	resp, err := s.client.Do(req, &response)
	if err != nil {
		return nil, resp, err
	}

	return &response, resp, nil
}

// Delete deletes a custom field from Snipe-IT.
func (s *FieldsService) Delete(id int) (*http.Response, error) {
	return s.DeleteContext(context.Background(), id)
}

// DeleteContext deletes a custom field from Snipe-IT with the provided context.
func (s *FieldsService) DeleteContext(ctx context.Context, id int) (*http.Response, error) {
	u := fmt.Sprintf("api/v1/fields/%d", id)
	req, err := s.client.newRequestWithContext(ctx, http.MethodDelete, u, nil)
	if err != nil {
		return nil, err
	}

	return s.client.Do(req, nil)
}

// Associate associates a custom field with a fieldset.
func (s *FieldsService) Associate(fieldID, fieldsetID int) (*http.Response, error) {
	return s.AssociateContext(context.Background(), fieldID, fieldsetID)
}

// AssociateContext associates a custom field with a fieldset with the provided context.
func (s *FieldsService) AssociateContext(ctx context.Context, fieldID, fieldsetID int) (*http.Response, error) {
	u := fmt.Sprintf("api/v1/fields/%d/associate", fieldID)
	body := map[string]int{"fieldset_id": fieldsetID}
	req, err := s.client.newRequestWithContext(ctx, http.MethodPost, u, body)
	if err != nil {
		return nil, err
	}

	return s.client.Do(req, nil)
}
