package snipeit

import (
	"context"
	"fmt"
	"net/http"
)

// SuppliersService handles communication with the supplier-related endpoints
// of the Snipe-IT API.
//
// Snipe-IT API docs: https://snipe-it.readme.io/reference/suppliers
type SuppliersService struct {
	client *Client
}

// SupplierResponse represents the API response for a single supplier.
type SupplierResponse struct {
	Response
	Payload Supplier `json:"payload"`
}

// SuppliersResponse represents the API response for multiple suppliers.
type SuppliersResponse struct {
	Response
	Rows []Supplier `json:"rows"`
}

// List returns a list of suppliers with pagination options.
func (s *SuppliersService) List(opts *ListOptions) (*SuppliersResponse, *http.Response, error) {
	return s.ListContext(context.Background(), opts)
}

// ListContext returns a list of suppliers with the provided context and pagination options.
func (s *SuppliersService) ListContext(ctx context.Context, opts *ListOptions) (*SuppliersResponse, *http.Response, error) {
	u := "api/v1/suppliers"
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

	var suppliers SuppliersResponse
	resp, err := s.client.Do(req, &suppliers)
	if err != nil {
		return nil, resp, err
	}

	return &suppliers, resp, nil
}

// Get fetches a single supplier by its ID.
func (s *SuppliersService) Get(id int) (*SupplierResponse, *http.Response, error) {
	return s.GetContext(context.Background(), id)
}

// GetContext fetches a single supplier by its ID with the provided context.
func (s *SuppliersService) GetContext(ctx context.Context, id int) (*SupplierResponse, *http.Response, error) {
	u := fmt.Sprintf("api/v1/suppliers/%d", id)
	req, err := s.client.newRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, nil, err
	}

	var supplier SupplierResponse
	resp, err := s.client.Do(req, &supplier)
	if err != nil {
		return nil, resp, err
	}

	return &supplier, resp, nil
}

// Create creates a new supplier in Snipe-IT.
func (s *SuppliersService) Create(supplier Supplier) (*SupplierResponse, *http.Response, error) {
	return s.CreateContext(context.Background(), supplier)
}

// CreateContext creates a new supplier in Snipe-IT with the provided context.
func (s *SuppliersService) CreateContext(ctx context.Context, supplier Supplier) (*SupplierResponse, *http.Response, error) {
	req, err := s.client.newRequestWithContext(ctx, http.MethodPost, "api/v1/suppliers", supplier)
	if err != nil {
		return nil, nil, err
	}

	var response SupplierResponse
	resp, err := s.client.Do(req, &response)
	if err != nil {
		return nil, resp, err
	}

	return &response, resp, nil
}

// Update updates an existing supplier in Snipe-IT.
func (s *SuppliersService) Update(id int, supplier Supplier) (*SupplierResponse, *http.Response, error) {
	return s.UpdateContext(context.Background(), id, supplier)
}

// UpdateContext updates an existing supplier in Snipe-IT with the provided context.
func (s *SuppliersService) UpdateContext(ctx context.Context, id int, supplier Supplier) (*SupplierResponse, *http.Response, error) {
	u := fmt.Sprintf("api/v1/suppliers/%d", id)
	req, err := s.client.newRequestWithContext(ctx, http.MethodPut, u, supplier)
	if err != nil {
		return nil, nil, err
	}

	var response SupplierResponse
	resp, err := s.client.Do(req, &response)
	if err != nil {
		return nil, resp, err
	}

	return &response, resp, nil
}

// Delete deletes a supplier from Snipe-IT.
func (s *SuppliersService) Delete(id int) (*http.Response, error) {
	return s.DeleteContext(context.Background(), id)
}

// DeleteContext deletes a supplier from Snipe-IT with the provided context.
func (s *SuppliersService) DeleteContext(ctx context.Context, id int) (*http.Response, error) {
	u := fmt.Sprintf("api/v1/suppliers/%d", id)
	req, err := s.client.newRequestWithContext(ctx, http.MethodDelete, u, nil)
	if err != nil {
		return nil, err
	}

	return s.client.Do(req, nil)
}
