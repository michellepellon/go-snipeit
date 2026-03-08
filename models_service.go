package snipeit

import (
	"context"
	"fmt"
	"net/http"
)

// ModelsService handles communication with the model-related endpoints
// of the Snipe-IT API.
//
// Snipe-IT API docs: https://snipe-it.readme.io/reference/models
type ModelsService struct {
	client *Client
}

// ModelResponse represents the API response for a single model.
type ModelResponse struct {
	Response
	Payload Model `json:"payload"`
}

// ModelsResponse represents the API response for multiple models.
type ModelsResponse struct {
	Response
	Rows []Model `json:"rows"`
}

// List returns a list of models with pagination options.
func (s *ModelsService) List(opts *ListOptions) (*ModelsResponse, *http.Response, error) {
	return s.ListContext(context.Background(), opts)
}

// ListContext returns a list of models with the provided context and pagination options.
func (s *ModelsService) ListContext(ctx context.Context, opts *ListOptions) (*ModelsResponse, *http.Response, error) {
	u := "api/v1/models"
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

	var models ModelsResponse
	resp, err := s.client.Do(req, &models)
	if err != nil {
		return nil, resp, err
	}

	return &models, resp, nil
}

// Get fetches a single model by its ID.
func (s *ModelsService) Get(id int) (*ModelResponse, *http.Response, error) {
	return s.GetContext(context.Background(), id)
}

// GetContext fetches a single model by its ID with the provided context.
func (s *ModelsService) GetContext(ctx context.Context, id int) (*ModelResponse, *http.Response, error) {
	u := fmt.Sprintf("api/v1/models/%d", id)
	req, err := s.client.newRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, nil, err
	}

	var model ModelResponse
	resp, err := s.client.Do(req, &model)
	if err != nil {
		return nil, resp, err
	}

	return &model, resp, nil
}

// Create creates a new model in Snipe-IT.
func (s *ModelsService) Create(model Model) (*ModelResponse, *http.Response, error) {
	return s.CreateContext(context.Background(), model)
}

// CreateContext creates a new model in Snipe-IT with the provided context.
func (s *ModelsService) CreateContext(ctx context.Context, model Model) (*ModelResponse, *http.Response, error) {
	req, err := s.client.newRequestWithContext(ctx, http.MethodPost, "api/v1/models", model)
	if err != nil {
		return nil, nil, err
	}

	var response ModelResponse
	resp, err := s.client.Do(req, &response)
	if err != nil {
		return nil, resp, err
	}

	return &response, resp, nil
}

// Update updates an existing model in Snipe-IT.
func (s *ModelsService) Update(id int, model Model) (*ModelResponse, *http.Response, error) {
	return s.UpdateContext(context.Background(), id, model)
}

// UpdateContext updates an existing model in Snipe-IT with the provided context.
func (s *ModelsService) UpdateContext(ctx context.Context, id int, model Model) (*ModelResponse, *http.Response, error) {
	u := fmt.Sprintf("api/v1/models/%d", id)
	req, err := s.client.newRequestWithContext(ctx, http.MethodPut, u, model)
	if err != nil {
		return nil, nil, err
	}

	var response ModelResponse
	resp, err := s.client.Do(req, &response)
	if err != nil {
		return nil, resp, err
	}

	return &response, resp, nil
}

// Delete deletes a model from Snipe-IT.
func (s *ModelsService) Delete(id int) (*http.Response, error) {
	return s.DeleteContext(context.Background(), id)
}

// DeleteContext deletes a model from Snipe-IT with the provided context.
func (s *ModelsService) DeleteContext(ctx context.Context, id int) (*http.Response, error) {
	u := fmt.Sprintf("api/v1/models/%d", id)
	req, err := s.client.newRequestWithContext(ctx, http.MethodDelete, u, nil)
	if err != nil {
		return nil, err
	}

	return s.client.Do(req, nil)
}
