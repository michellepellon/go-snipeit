package snipeit

import (
	"context"
	"fmt"
	"net/http"
)

// LocationsService handles communication with the location-related endpoints
// of the Snipe-IT API.
//
// Snipe-IT API docs: https://snipe-it.readme.io/reference/locations
type LocationsService struct {
	client *Client
}

// LocationResponse represents the API response for a single location.
type LocationResponse struct {
	Response
	Payload Location `json:"payload"`
}

// LocationsResponse represents the API response for multiple locations.
type LocationsResponse struct {
	Response
	Rows []Location `json:"rows"`
}

// List returns a list of locations with pagination options.
func (s *LocationsService) List(opts *ListOptions) (*LocationsResponse, *http.Response, error) {
	return s.ListContext(context.Background(), opts)
}

// ListContext returns a list of locations with the provided context and pagination options.
func (s *LocationsService) ListContext(ctx context.Context, opts *ListOptions) (*LocationsResponse, *http.Response, error) {
	u := "api/v1/locations"
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

	var locations LocationsResponse
	resp, err := s.client.Do(req, &locations)
	if err != nil {
		return nil, resp, err
	}

	return &locations, resp, nil
}

// Get fetches a single location by its ID.
func (s *LocationsService) Get(id int) (*LocationResponse, *http.Response, error) {
	return s.GetContext(context.Background(), id)
}

// GetContext fetches a single location by its ID with the provided context.
func (s *LocationsService) GetContext(ctx context.Context, id int) (*LocationResponse, *http.Response, error) {
	u := fmt.Sprintf("api/v1/locations/%d", id)
	req, err := s.client.newRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, nil, err
	}

	var location LocationResponse
	resp, err := s.client.Do(req, &location)
	if err != nil {
		return nil, resp, err
	}

	return &location, resp, nil
}

// Create creates a new location in Snipe-IT.
func (s *LocationsService) Create(location Location) (*LocationResponse, *http.Response, error) {
	return s.CreateContext(context.Background(), location)
}

// CreateContext creates a new location in Snipe-IT with the provided context.
func (s *LocationsService) CreateContext(ctx context.Context, location Location) (*LocationResponse, *http.Response, error) {
	req, err := s.client.newRequestWithContext(ctx, http.MethodPost, "api/v1/locations", location)
	if err != nil {
		return nil, nil, err
	}

	var response LocationResponse
	resp, err := s.client.Do(req, &response)
	if err != nil {
		return nil, resp, err
	}

	return &response, resp, nil
}

// Update updates an existing location in Snipe-IT.
func (s *LocationsService) Update(id int, location Location) (*LocationResponse, *http.Response, error) {
	return s.UpdateContext(context.Background(), id, location)
}

// UpdateContext updates an existing location in Snipe-IT with the provided context.
func (s *LocationsService) UpdateContext(ctx context.Context, id int, location Location) (*LocationResponse, *http.Response, error) {
	u := fmt.Sprintf("api/v1/locations/%d", id)
	req, err := s.client.newRequestWithContext(ctx, http.MethodPut, u, location)
	if err != nil {
		return nil, nil, err
	}

	var response LocationResponse
	resp, err := s.client.Do(req, &response)
	if err != nil {
		return nil, resp, err
	}

	return &response, resp, nil
}

// Delete deletes a location from Snipe-IT.
func (s *LocationsService) Delete(id int) (*http.Response, error) {
	return s.DeleteContext(context.Background(), id)
}

// DeleteContext deletes a location from Snipe-IT with the provided context.
func (s *LocationsService) DeleteContext(ctx context.Context, id int) (*http.Response, error) {
	u := fmt.Sprintf("api/v1/locations/%d", id)
	req, err := s.client.newRequestWithContext(ctx, http.MethodDelete, u, nil)
	if err != nil {
		return nil, err
	}

	return s.client.Do(req, nil)
}
