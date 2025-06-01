// Package snipeit provides a client for the Snipe-IT Asset Management API.
package snipeit

import (
	"context"
	"fmt"
	"net/http"
)

// AssetsService handles communication with the asset-related endpoints
// of the Snipe-IT API.
//
// Snipe-IT API docs: https://snipe-it.readme.io/reference/hardware
type AssetsService struct {
	client *Client
}

// AssetResponse represents the API response for a single asset.
// It embeds the standard Response struct and adds a Payload field
// that contains the Asset data.
type AssetResponse struct {
	Response
	// Payload contains the actual Asset data
	Payload Asset `json:"payload"`
}

// AssetsResponse represents the API response for multiple assets.
// It embeds the standard Response struct and adds a Rows field
// that contains a slice of Assets.
type AssetsResponse struct {
	Response
	// Rows contains the list of Asset objects
	Rows []Asset `json:"rows"`
}

// List returns a list of assets with pagination options.
//
// opts can be used to customize the response with pagination, search, and sorting.
// If opts is nil, default pagination values will be used.
//
// Snipe-IT API docs: https://snipe-it.readme.io/reference/hardware-list
func (s *AssetsService) List(opts *ListOptions) (*AssetsResponse, *http.Response, error) {
	return s.ListContext(context.Background(), opts)
}

// ListContext returns a list of assets with the provided context and pagination options.
//
// ctx is the context for the request.
// opts can be used to customize the response with pagination, search, and sorting.
// If opts is nil, default pagination values will be used.
//
// Snipe-IT API docs: https://snipe-it.readme.io/reference/hardware-list
func (s *AssetsService) ListContext(ctx context.Context, opts *ListOptions) (*AssetsResponse, *http.Response, error) {
	u := "api/v1/hardware"
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

	var assets AssetsResponse
	resp, err := s.client.Do(req, &assets)
	if err != nil {
		return nil, resp, err
	}

	return &assets, resp, nil
}

// Get fetches a single asset by its ID.
//
// id is the unique identifier of the asset to retrieve.
//
// Snipe-IT API docs: https://snipe-it.readme.io/reference/hardware-by-id
func (s *AssetsService) Get(id int) (*AssetResponse, *http.Response, error) {
	return s.GetContext(context.Background(), id)
}

// GetContext fetches a single asset by its ID with the provided context.
//
// ctx is the context for the request.
// id is the unique identifier of the asset to retrieve.
//
// Snipe-IT API docs: https://snipe-it.readme.io/reference/hardware-by-id
func (s *AssetsService) GetContext(ctx context.Context, id int) (*AssetResponse, *http.Response, error) {
	u := fmt.Sprintf("api/v1/hardware/%d", id)
	req, err := s.client.newRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, nil, err
	}

	var asset AssetResponse
	resp, err := s.client.Do(req, &asset)
	if err != nil {
		return nil, resp, err
	}

	return &asset, resp, nil
}

// Create creates a new asset in Snipe-IT.
//
// asset must contain the required fields:
// - Model.ID: The ID of the model
// - StatusLabel.ID: The ID of the status label
// - AssetTag: A unique asset tag
//
// Other important fields that should be considered:
// - Name: A name for the asset
// - Serial: The manufacturer's serial number
//
// Snipe-IT API docs: https://snipe-it.readme.io/reference/hardware-create
func (s *AssetsService) Create(asset Asset) (*AssetResponse, *http.Response, error) {
	return s.CreateContext(context.Background(), asset)
}

// CreateContext creates a new asset in Snipe-IT with the provided context.
//
// ctx is the context for the request.
// asset must contain the required fields:
// - Model.ID: The ID of the model
// - StatusLabel.ID: The ID of the status label
// - AssetTag: A unique asset tag
//
// Other important fields that should be considered:
// - Name: A name for the asset
// - Serial: The manufacturer's serial number
//
// Snipe-IT API docs: https://snipe-it.readme.io/reference/hardware-create
func (s *AssetsService) CreateContext(ctx context.Context, asset Asset) (*AssetResponse, *http.Response, error) {
	req, err := s.client.newRequestWithContext(ctx, http.MethodPost, "api/v1/hardware", asset)
	if err != nil {
		return nil, nil, err
	}

	var response AssetResponse
	resp, err := s.client.Do(req, &response)
	if err != nil {
		return nil, resp, err
	}

	return &response, resp, nil
}

// Update updates an existing asset in Snipe-IT.
//
// id is the unique identifier of the asset to update.
// asset contains the fields to update. You only need to include
// the fields you want to modify; other fields can be omitted.
//
// Snipe-IT API docs: https://snipe-it.readme.io/reference/hardware-update
func (s *AssetsService) Update(id int, asset Asset) (*AssetResponse, *http.Response, error) {
	return s.UpdateContext(context.Background(), id, asset)
}

// UpdateContext updates an existing asset in Snipe-IT with the provided context.
//
// ctx is the context for the request.
// id is the unique identifier of the asset to update.
// asset contains the fields to update. You only need to include
// the fields you want to modify; other fields can be omitted.
//
// Snipe-IT API docs: https://snipe-it.readme.io/reference/hardware-update
func (s *AssetsService) UpdateContext(ctx context.Context, id int, asset Asset) (*AssetResponse, *http.Response, error) {
	u := fmt.Sprintf("api/v1/hardware/%d", id)
	req, err := s.client.newRequestWithContext(ctx, http.MethodPut, u, asset)
	if err != nil {
		return nil, nil, err
	}

	var response AssetResponse
	resp, err := s.client.Do(req, &response)
	if err != nil {
		return nil, resp, err
	}

	return &response, resp, nil
}

// Delete deletes an asset from Snipe-IT.
//
// id is the unique identifier of the asset to delete.
// This operation soft-deletes the asset in Snipe-IT.
//
// Snipe-IT API docs: https://snipe-it.readme.io/reference/hardware-delete
func (s *AssetsService) Delete(id int) (*http.Response, error) {
	return s.DeleteContext(context.Background(), id)
}

// DeleteContext deletes an asset from Snipe-IT with the provided context.
//
// ctx is the context for the request.
// id is the unique identifier of the asset to delete.
// This operation soft-deletes the asset in Snipe-IT.
//
// Snipe-IT API docs: https://snipe-it.readme.io/reference/hardware-delete
func (s *AssetsService) DeleteContext(ctx context.Context, id int) (*http.Response, error) {
	u := fmt.Sprintf("api/v1/hardware/%d", id)
	req, err := s.client.newRequestWithContext(ctx, http.MethodDelete, u, nil)
	if err != nil {
		return nil, err
	}

	return s.client.Do(req, nil)
}

// Checkout assigns an asset to a user, location, or another asset.
//
// id is the unique identifier of the asset to check out.
// checkout is a map containing checkout parameters, such as:
// - assigned_user: ID of the user to assign the asset to
// - assigned_asset: ID of the asset to assign this asset to
// - assigned_location: ID of the location to assign the asset to
// - checkout_at: Date of checkout (YYYY-MM-DD format)
// - expected_checkin: Expected checkin date (YYYY-MM-DD format)
// - note: Note about the checkout
//
// Snipe-IT API docs: https://snipe-it.readme.io/reference/hardware-checkout
func (s *AssetsService) Checkout(id int, checkout map[string]interface{}) (*AssetResponse, *http.Response, error) {
	return s.CheckoutContext(context.Background(), id, checkout)
}

// CheckoutContext assigns an asset to a user, location, or another asset with the provided context.
//
// ctx is the context for the request.
// id is the unique identifier of the asset to check out.
// checkout is a map containing checkout parameters, such as:
// - assigned_user: ID of the user to assign the asset to
// - assigned_asset: ID of the asset to assign this asset to
// - assigned_location: ID of the location to assign the asset to
// - checkout_at: Date of checkout (YYYY-MM-DD format)
// - expected_checkin: Expected checkin date (YYYY-MM-DD format)
// - note: Note about the checkout
//
// Snipe-IT API docs: https://snipe-it.readme.io/reference/hardware-checkout
func (s *AssetsService) CheckoutContext(ctx context.Context, id int, checkout map[string]interface{}) (*AssetResponse, *http.Response, error) {
	u := fmt.Sprintf("api/v1/hardware/%d/checkout", id)
	req, err := s.client.newRequestWithContext(ctx, http.MethodPost, u, checkout)
	if err != nil {
		return nil, nil, err
	}

	var response AssetResponse
	resp, err := s.client.Do(req, &response)
	if err != nil {
		return nil, resp, err
	}

	return &response, resp, nil
}

// Checkin returns an asset from a user, location, or asset it was assigned to.
//
// id is the unique identifier of the asset to check in.
// checkin is a map containing checkin parameters, such as:
// - note: Note about the checkin
// - location_id: ID of the location to assign after checkin
// - status_id: ID of the status to assign after checkin
// - checkin_at: Date of checkin (YYYY-MM-DD format)
//
// Snipe-IT API docs: https://snipe-it.readme.io/reference/hardware-checkin
func (s *AssetsService) Checkin(id int, checkin map[string]interface{}) (*AssetResponse, *http.Response, error) {
	return s.CheckinContext(context.Background(), id, checkin)
}

// CheckinContext returns an asset with the provided context.
//
// ctx is the context for the request.
// id is the unique identifier of the asset to check in.
// checkin is a map containing checkin parameters, such as:
// - note: Note about the checkin
// - location_id: ID of the location to assign after checkin
// - status_id: ID of the status to assign after checkin
// - checkin_at: Date of checkin (YYYY-MM-DD format)
//
// Snipe-IT API docs: https://snipe-it.readme.io/reference/hardware-checkin
func (s *AssetsService) CheckinContext(ctx context.Context, id int, checkin map[string]interface{}) (*AssetResponse, *http.Response, error) {
	u := fmt.Sprintf("api/v1/hardware/%d/checkin", id)
	req, err := s.client.newRequestWithContext(ctx, http.MethodPost, u, checkin)
	if err != nil {
		return nil, nil, err
	}

	var response AssetResponse
	resp, err := s.client.Do(req, &response)
	if err != nil {
		return nil, resp, err
	}

	return &response, resp, nil
}

// GetAssetBySerial fetches a single asset by its serial number.
//
// serial is the manufacturer's serial number of the asset to retrieve.
//
// Snipe-IT API docs: https://snipe-it.readme.io/reference/hardware-by-serial
func (s *AssetsService) GetAssetBySerial(serial string) (*AssetResponse, *http.Response, error) {
	return s.GetAssetBySerialContext(context.Background(), serial)
}

// GetAssetBySerialContext fetches a single asset by its serial number with the provided context.
//
// ctx is the context for the request.
// serial is the manufacturer's serial number of the asset to retrieve.
//
// Snipe-IT API docs: https://snipe-it.readme.io/reference/hardware-by-serial
func (s *AssetsService) GetAssetBySerialContext(ctx context.Context, serial string) (*AssetResponse, *http.Response, error) {
	u := fmt.Sprintf("api/v1/hardware/byserial/%s", serial)
	req, err := s.client.newRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, nil, err
	}

	var asset AssetResponse
	resp, err := s.client.Do(req, &asset)
	if err != nil {
		return nil, resp, err
	}

	return &asset, resp, nil
}