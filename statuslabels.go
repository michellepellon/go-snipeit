package snipeit

import (
	"context"
	"fmt"
	"net/http"
)

// StatusLabelsService handles communication with the status label-related endpoints
// of the Snipe-IT API.
//
// Snipe-IT API docs: https://snipe-it.readme.io/reference/statuslabels
type StatusLabelsService struct {
	client *Client
}

// StatusLabelResponse represents the API response for a single status label.
type StatusLabelResponse struct {
	Response
	Payload StatusLabel `json:"payload"`
}

// StatusLabelsResponse represents the API response for multiple status labels.
type StatusLabelsResponse struct {
	Response
	Rows []StatusLabel `json:"rows"`
}

// List returns a list of status labels with pagination options.
func (s *StatusLabelsService) List(opts *ListOptions) (*StatusLabelsResponse, *http.Response, error) {
	return s.ListContext(context.Background(), opts)
}

// ListContext returns a list of status labels with the provided context and pagination options.
func (s *StatusLabelsService) ListContext(ctx context.Context, opts *ListOptions) (*StatusLabelsResponse, *http.Response, error) {
	u := "api/v1/statuslabels"
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

	var labels StatusLabelsResponse
	resp, err := s.client.Do(req, &labels)
	if err != nil {
		return nil, resp, err
	}

	return &labels, resp, nil
}

// Get fetches a single status label by its ID.
func (s *StatusLabelsService) Get(id int) (*StatusLabelResponse, *http.Response, error) {
	return s.GetContext(context.Background(), id)
}

// GetContext fetches a single status label by its ID with the provided context.
func (s *StatusLabelsService) GetContext(ctx context.Context, id int) (*StatusLabelResponse, *http.Response, error) {
	u := fmt.Sprintf("api/v1/statuslabels/%d", id)
	req, err := s.client.newRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, nil, err
	}

	var label StatusLabelResponse
	resp, err := s.client.Do(req, &label)
	if err != nil {
		return nil, resp, err
	}

	return &label, resp, nil
}

// Create creates a new status label in Snipe-IT.
func (s *StatusLabelsService) Create(label StatusLabel) (*StatusLabelResponse, *http.Response, error) {
	return s.CreateContext(context.Background(), label)
}

// CreateContext creates a new status label in Snipe-IT with the provided context.
func (s *StatusLabelsService) CreateContext(ctx context.Context, label StatusLabel) (*StatusLabelResponse, *http.Response, error) {
	req, err := s.client.newRequestWithContext(ctx, http.MethodPost, "api/v1/statuslabels", label)
	if err != nil {
		return nil, nil, err
	}

	var response StatusLabelResponse
	resp, err := s.client.Do(req, &response)
	if err != nil {
		return nil, resp, err
	}

	return &response, resp, nil
}

// Update updates an existing status label in Snipe-IT.
func (s *StatusLabelsService) Update(id int, label StatusLabel) (*StatusLabelResponse, *http.Response, error) {
	return s.UpdateContext(context.Background(), id, label)
}

// UpdateContext updates an existing status label in Snipe-IT with the provided context.
func (s *StatusLabelsService) UpdateContext(ctx context.Context, id int, label StatusLabel) (*StatusLabelResponse, *http.Response, error) {
	u := fmt.Sprintf("api/v1/statuslabels/%d", id)
	req, err := s.client.newRequestWithContext(ctx, http.MethodPut, u, label)
	if err != nil {
		return nil, nil, err
	}

	var response StatusLabelResponse
	resp, err := s.client.Do(req, &response)
	if err != nil {
		return nil, resp, err
	}

	return &response, resp, nil
}

// Delete deletes a status label from Snipe-IT.
func (s *StatusLabelsService) Delete(id int) (*http.Response, error) {
	return s.DeleteContext(context.Background(), id)
}

// DeleteContext deletes a status label from Snipe-IT with the provided context.
func (s *StatusLabelsService) DeleteContext(ctx context.Context, id int) (*http.Response, error) {
	u := fmt.Sprintf("api/v1/statuslabels/%d", id)
	req, err := s.client.newRequestWithContext(ctx, http.MethodDelete, u, nil)
	if err != nil {
		return nil, err
	}

	return s.client.Do(req, nil)
}
