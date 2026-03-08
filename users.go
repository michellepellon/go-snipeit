package snipeit

import (
	"context"
	"fmt"
	"net/http"
)

// UsersService handles communication with the user-related endpoints
// of the Snipe-IT API.
//
// Snipe-IT API docs: https://snipe-it.readme.io/reference/users
type UsersService struct {
	client *Client
}

// UserResponse represents the API response for a single user.
type UserResponse struct {
	Response
	Payload User `json:"payload"`
}

// UsersResponse represents the API response for multiple users.
type UsersResponse struct {
	Response
	Rows []User `json:"rows"`
}

// List returns a list of users with pagination options.
func (s *UsersService) List(opts *ListOptions) (*UsersResponse, *http.Response, error) {
	return s.ListContext(context.Background(), opts)
}

// ListContext returns a list of users with the provided context and pagination options.
func (s *UsersService) ListContext(ctx context.Context, opts *ListOptions) (*UsersResponse, *http.Response, error) {
	u := "api/v1/users"
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

	var users UsersResponse
	resp, err := s.client.Do(req, &users)
	if err != nil {
		return nil, resp, err
	}

	return &users, resp, nil
}

// Get fetches a single user by their ID.
func (s *UsersService) Get(id int) (*UserResponse, *http.Response, error) {
	return s.GetContext(context.Background(), id)
}

// GetContext fetches a single user by their ID with the provided context.
func (s *UsersService) GetContext(ctx context.Context, id int) (*UserResponse, *http.Response, error) {
	u := fmt.Sprintf("api/v1/users/%d", id)
	req, err := s.client.newRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, nil, err
	}

	var user UserResponse
	resp, err := s.client.Do(req, &user)
	if err != nil {
		return nil, resp, err
	}

	return &user, resp, nil
}

// Create creates a new user in Snipe-IT.
func (s *UsersService) Create(user User) (*UserResponse, *http.Response, error) {
	return s.CreateContext(context.Background(), user)
}

// CreateContext creates a new user in Snipe-IT with the provided context.
func (s *UsersService) CreateContext(ctx context.Context, user User) (*UserResponse, *http.Response, error) {
	req, err := s.client.newRequestWithContext(ctx, http.MethodPost, "api/v1/users", user)
	if err != nil {
		return nil, nil, err
	}

	var response UserResponse
	resp, err := s.client.Do(req, &response)
	if err != nil {
		return nil, resp, err
	}

	return &response, resp, nil
}

// Update updates an existing user in Snipe-IT.
func (s *UsersService) Update(id int, user User) (*UserResponse, *http.Response, error) {
	return s.UpdateContext(context.Background(), id, user)
}

// UpdateContext updates an existing user in Snipe-IT with the provided context.
func (s *UsersService) UpdateContext(ctx context.Context, id int, user User) (*UserResponse, *http.Response, error) {
	u := fmt.Sprintf("api/v1/users/%d", id)
	req, err := s.client.newRequestWithContext(ctx, http.MethodPut, u, user)
	if err != nil {
		return nil, nil, err
	}

	var response UserResponse
	resp, err := s.client.Do(req, &response)
	if err != nil {
		return nil, resp, err
	}

	return &response, resp, nil
}

// Delete deletes a user from Snipe-IT.
func (s *UsersService) Delete(id int) (*http.Response, error) {
	return s.DeleteContext(context.Background(), id)
}

// DeleteContext deletes a user from Snipe-IT with the provided context.
func (s *UsersService) DeleteContext(ctx context.Context, id int) (*http.Response, error) {
	u := fmt.Sprintf("api/v1/users/%d", id)
	req, err := s.client.newRequestWithContext(ctx, http.MethodDelete, u, nil)
	if err != nil {
		return nil, err
	}

	return s.client.Do(req, nil)
}
