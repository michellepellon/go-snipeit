package snipeit

import (
	"context"
	"fmt"
	"net/http"
)

// CategoriesService handles communication with the category-related endpoints
// of the Snipe-IT API.
//
// Snipe-IT API docs: https://snipe-it.readme.io/reference/categories
type CategoriesService struct {
	client *Client
}

// CategoryResponse represents the API response for a single category.
type CategoryResponse struct {
	Response
	Payload Category `json:"payload"`
}

// CategoriesResponse represents the API response for multiple categories.
type CategoriesResponse struct {
	Response
	Rows []Category `json:"rows"`
}

// List returns a list of categories with pagination options.
func (s *CategoriesService) List(opts *ListOptions) (*CategoriesResponse, *http.Response, error) {
	return s.ListContext(context.Background(), opts)
}

// ListContext returns a list of categories with the provided context and pagination options.
func (s *CategoriesService) ListContext(ctx context.Context, opts *ListOptions) (*CategoriesResponse, *http.Response, error) {
	u := "api/v1/categories"
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

	var categories CategoriesResponse
	resp, err := s.client.Do(req, &categories)
	if err != nil {
		return nil, resp, err
	}

	return &categories, resp, nil
}

// Get fetches a single category by its ID.
func (s *CategoriesService) Get(id int) (*CategoryResponse, *http.Response, error) {
	return s.GetContext(context.Background(), id)
}

// GetContext fetches a single category by its ID with the provided context.
func (s *CategoriesService) GetContext(ctx context.Context, id int) (*CategoryResponse, *http.Response, error) {
	u := fmt.Sprintf("api/v1/categories/%d", id)
	req, err := s.client.newRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, nil, err
	}

	var category CategoryResponse
	resp, err := s.client.Do(req, &category)
	if err != nil {
		return nil, resp, err
	}

	return &category, resp, nil
}

// Create creates a new category in Snipe-IT.
func (s *CategoriesService) Create(category Category) (*CategoryResponse, *http.Response, error) {
	return s.CreateContext(context.Background(), category)
}

// CreateContext creates a new category in Snipe-IT with the provided context.
func (s *CategoriesService) CreateContext(ctx context.Context, category Category) (*CategoryResponse, *http.Response, error) {
	req, err := s.client.newRequestWithContext(ctx, http.MethodPost, "api/v1/categories", category)
	if err != nil {
		return nil, nil, err
	}

	var response CategoryResponse
	resp, err := s.client.Do(req, &response)
	if err != nil {
		return nil, resp, err
	}

	return &response, resp, nil
}

// Update updates an existing category in Snipe-IT.
func (s *CategoriesService) Update(id int, category Category) (*CategoryResponse, *http.Response, error) {
	return s.UpdateContext(context.Background(), id, category)
}

// UpdateContext updates an existing category in Snipe-IT with the provided context.
func (s *CategoriesService) UpdateContext(ctx context.Context, id int, category Category) (*CategoryResponse, *http.Response, error) {
	u := fmt.Sprintf("api/v1/categories/%d", id)
	req, err := s.client.newRequestWithContext(ctx, http.MethodPut, u, category)
	if err != nil {
		return nil, nil, err
	}

	var response CategoryResponse
	resp, err := s.client.Do(req, &response)
	if err != nil {
		return nil, resp, err
	}

	return &response, resp, nil
}

// Delete deletes a category from Snipe-IT.
func (s *CategoriesService) Delete(id int) (*http.Response, error) {
	return s.DeleteContext(context.Background(), id)
}

// DeleteContext deletes a category from Snipe-IT with the provided context.
func (s *CategoriesService) DeleteContext(ctx context.Context, id int) (*http.Response, error) {
	u := fmt.Sprintf("api/v1/categories/%d", id)
	req, err := s.client.newRequestWithContext(ctx, http.MethodDelete, u, nil)
	if err != nil {
		return nil, err
	}

	return s.client.Do(req, nil)
}
