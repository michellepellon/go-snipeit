package snipeit

import (
	"context"
	"fmt"
	"net/http"
)

// ManufacturersService handles communication with the manufacturer-related endpoints
// of the Snipe-IT API.
//
// Snipe-IT API docs: https://snipe-it.readme.io/reference/manufacturers
type ManufacturersService struct {
	client *Client
}

// ManufacturerResponse represents the API response for a single manufacturer.
type ManufacturerResponse struct {
	Response
	Payload Manufacturer `json:"payload"`
}

// ManufacturersResponse represents the API response for multiple manufacturers.
type ManufacturersResponse struct {
	Response
	Rows []Manufacturer `json:"rows"`
}

// List returns a list of manufacturers with pagination options.
func (s *ManufacturersService) List(opts *ListOptions) (*ManufacturersResponse, *http.Response, error) {
	return s.ListContext(context.Background(), opts)
}

// ListContext returns a list of manufacturers with the provided context and pagination options.
func (s *ManufacturersService) ListContext(ctx context.Context, opts *ListOptions) (*ManufacturersResponse, *http.Response, error) {
	u := "api/v1/manufacturers"
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

	var manufacturers ManufacturersResponse
	resp, err := s.client.Do(req, &manufacturers)
	if err != nil {
		return nil, resp, err
	}

	return &manufacturers, resp, nil
}

// Get fetches a single manufacturer by its ID.
func (s *ManufacturersService) Get(id int) (*ManufacturerResponse, *http.Response, error) {
	return s.GetContext(context.Background(), id)
}

// GetContext fetches a single manufacturer by its ID with the provided context.
func (s *ManufacturersService) GetContext(ctx context.Context, id int) (*ManufacturerResponse, *http.Response, error) {
	u := fmt.Sprintf("api/v1/manufacturers/%d", id)
	req, err := s.client.newRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, nil, err
	}

	var manufacturer ManufacturerResponse
	resp, err := s.client.Do(req, &manufacturer)
	if err != nil {
		return nil, resp, err
	}

	return &manufacturer, resp, nil
}

// Create creates a new manufacturer in Snipe-IT.
func (s *ManufacturersService) Create(manufacturer Manufacturer) (*ManufacturerResponse, *http.Response, error) {
	return s.CreateContext(context.Background(), manufacturer)
}

// CreateContext creates a new manufacturer in Snipe-IT with the provided context.
func (s *ManufacturersService) CreateContext(ctx context.Context, manufacturer Manufacturer) (*ManufacturerResponse, *http.Response, error) {
	req, err := s.client.newRequestWithContext(ctx, http.MethodPost, "api/v1/manufacturers", manufacturer)
	if err != nil {
		return nil, nil, err
	}

	var response ManufacturerResponse
	resp, err := s.client.Do(req, &response)
	if err != nil {
		return nil, resp, err
	}

	return &response, resp, nil
}

// Update updates an existing manufacturer in Snipe-IT.
func (s *ManufacturersService) Update(id int, manufacturer Manufacturer) (*ManufacturerResponse, *http.Response, error) {
	return s.UpdateContext(context.Background(), id, manufacturer)
}

// UpdateContext updates an existing manufacturer in Snipe-IT with the provided context.
func (s *ManufacturersService) UpdateContext(ctx context.Context, id int, manufacturer Manufacturer) (*ManufacturerResponse, *http.Response, error) {
	u := fmt.Sprintf("api/v1/manufacturers/%d", id)
	req, err := s.client.newRequestWithContext(ctx, http.MethodPut, u, manufacturer)
	if err != nil {
		return nil, nil, err
	}

	var response ManufacturerResponse
	resp, err := s.client.Do(req, &response)
	if err != nil {
		return nil, resp, err
	}

	return &response, resp, nil
}

// Delete deletes a manufacturer from Snipe-IT.
func (s *ManufacturersService) Delete(id int) (*http.Response, error) {
	return s.DeleteContext(context.Background(), id)
}

// DeleteContext deletes a manufacturer from Snipe-IT with the provided context.
func (s *ManufacturersService) DeleteContext(ctx context.Context, id int) (*http.Response, error) {
	u := fmt.Sprintf("api/v1/manufacturers/%d", id)
	req, err := s.client.newRequestWithContext(ctx, http.MethodDelete, u, nil)
	if err != nil {
		return nil, err
	}

	return s.client.Do(req, nil)
}
