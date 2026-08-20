package snipeit

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// LicensesService handles communication with the license- and seat-related
// endpoints of the Snipe-IT API.
//
// Snipe-IT API docs: https://snipe-it.readme.io/reference/licenses
type LicensesService struct {
	client *Client
}

// License represents a Snipe-IT software license.
//
// Like Asset, a License reads back with nested objects (category, manufacturer,
// supplier) but is written with flat *_id fields; MarshalJSON handles the
// conversion, so the same struct serves both directions.
type License struct {
	CommonFields

	// Seats is the total number of seats on the license.
	//
	// Snipe-IT limits how much a single create or update may change this value
	// (10,000 by default, License::prepareLimitChangeRule) relative to the
	// license's current seat count, so growing far past that takes repeated
	// updates rather than one large one.
	Seats int `json:"seats"`

	// FreeSeatsCount is the number of seats not currently checked out.
	FreeSeatsCount int `json:"free_seats_count"`

	// Category is the license's category as returned by the API. Snipe-IT
	// requires it to be a category of type "license".
	Category *Category `json:"category,omitempty"`

	// CategoryID sets the category on create/update.
	CategoryID int `json:"-"`

	// Manufacturer of the licensed software, as returned by the API.
	Manufacturer *Manufacturer `json:"manufacturer,omitempty"`

	// ManufacturerID sets the manufacturer on create/update.
	ManufacturerID int `json:"-"`

	// Supplier the license was purchased from, as returned by the API.
	Supplier *Supplier `json:"supplier,omitempty"`

	// SupplierID sets the supplier on create/update.
	SupplierID int `json:"-"`

	// CompanyID scopes the license to a company in full-company-support setups.
	CompanyID int `json:"company_id,omitempty"`

	// Reassignable reports whether a seat may be checked in and reissued.
	Reassignable FlexBool `json:"reassignable"`

	// ProductKey is the license key.
	ProductKey string `json:"product_key,omitempty"`

	// LicensedToName and LicensedToEmail record who the license is issued to.
	LicensedToName  string `json:"license_name,omitempty"`
	LicensedToEmail string `json:"license_email,omitempty"`

	// OrderNumber and PurchaseOrder record the purchase paperwork.
	OrderNumber   string `json:"order_number,omitempty"`
	PurchaseOrder string `json:"purchase_order,omitempty"`

	// PurchaseCost is the cost of the license. The API returns it as a
	// formatted string, so it is written back the same way (e.g. "49.99").
	PurchaseCost string `json:"purchase_cost,omitempty"`

	// PurchaseDate is when the license was purchased.
	PurchaseDate *SnipeTime `json:"purchase_date,omitempty"`

	// ExpirationDate is when the license expires.
	//
	// On write, nil leaves the stored value untouched while a non-nil zero
	// time clears it — Snipe-IT only drops an existing expiration when the
	// field is sent explicitly as null.
	ExpirationDate *SnipeTime `json:"expiration_date,omitempty"`

	// Notes on the license (also present on CommonFields for reads).
	Maintained FlexBool `json:"maintained,omitempty"`
}

// MarshalJSON implements json.Marshaler for License. The Snipe-IT API returns
// nested objects for related resources but expects flat ID fields on write, and
// dates as "YYYY-MM-DD". A nil date field is omitted; a non-nil zero date is
// sent as null so an existing value is cleared.
func (l License) MarshalJSON() ([]byte, error) {
	m := make(map[string]interface{})

	if l.Name != "" {
		m["name"] = l.Name
	}
	if l.Seats != 0 {
		m["seats"] = l.Seats
	}
	if id := l.categoryID(); id != 0 {
		m["category_id"] = id
	}
	if id := l.manufacturerID(); id != 0 {
		m["manufacturer_id"] = id
	}
	if id := l.supplierID(); id != 0 {
		m["supplier_id"] = id
	}
	if l.CompanyID != 0 {
		m["company_id"] = l.CompanyID
	}
	m["reassignable"] = bool(l.Reassignable)
	if l.ProductKey != "" {
		m["product_key"] = l.ProductKey
	}
	if l.LicensedToName != "" {
		m["license_name"] = l.LicensedToName
	}
	if l.LicensedToEmail != "" {
		m["license_email"] = l.LicensedToEmail
	}
	if l.OrderNumber != "" {
		m["order_number"] = l.OrderNumber
	}
	if l.PurchaseOrder != "" {
		m["purchase_order"] = l.PurchaseOrder
	}
	if l.PurchaseCost != "" {
		m["purchase_cost"] = l.PurchaseCost
	}
	if l.Notes != "" {
		m["notes"] = l.Notes
	}
	if d, ok := writeDate(l.PurchaseDate); ok {
		m["purchase_date"] = d
	}
	if d, ok := writeDate(l.ExpirationDate); ok {
		m["expiration_date"] = d
	}

	return json.Marshal(m)
}

// writeDate renders a date field for the write API: nil is omitted (ok=false),
// a zero time is sent as null to clear the stored value, and anything else is
// sent in Snipe-IT's "YYYY-MM-DD" form.
func writeDate(t *SnipeTime) (interface{}, bool) {
	if t == nil {
		return nil, false
	}
	if t.IsZero() {
		return nil, true
	}
	return t.Format("2006-01-02"), true
}

// categoryID prefers the flat ID and falls back to the nested object, so a
// license read from the API can be written back unchanged.
func (l License) categoryID() int {
	if l.CategoryID != 0 {
		return l.CategoryID
	}
	if l.Category != nil {
		return l.Category.ID
	}
	return 0
}

func (l License) manufacturerID() int {
	if l.ManufacturerID != 0 {
		return l.ManufacturerID
	}
	if l.Manufacturer != nil {
		return l.Manufacturer.ID
	}
	return 0
}

func (l License) supplierID() int {
	if l.SupplierID != 0 {
		return l.SupplierID
	}
	if l.Supplier != nil {
		return l.Supplier.ID
	}
	return 0
}

// LicenseSeat represents a single seat on a license. A seat is checked out to
// at most one of AssignedUser or AssignedAsset; both are nil when it is free.
type LicenseSeat struct {
	// ID is the seat's id, used in the seat endpoints' path.
	ID int `json:"id"`

	// LicenseID is the id of the license the seat belongs to.
	LicenseID int `json:"license_id"`

	// AssignedUser is the user holding the seat, nil when unassigned.
	AssignedUser *User `json:"assigned_user,omitempty"`

	// AssignedAsset is the asset holding the seat, nil when unassigned.
	AssignedAsset *Asset `json:"assigned_asset,omitempty"`

	// Reassignable mirrors the parent license's reassignable flag.
	Reassignable FlexBool `json:"reassignable"`
}

// LicenseResponse represents the API response for a single license.
type LicenseResponse struct {
	Response
	Payload License `json:"payload"`
}

// LicensesResponse represents the API response for multiple licenses.
type LicensesResponse struct {
	Response
	Rows []License `json:"rows"`
}

// LicenseSeatsResponse represents the API response for a license's seats.
type LicenseSeatsResponse struct {
	Response
	Rows []LicenseSeat `json:"rows"`
}

// List returns a list of licenses with pagination options.
func (s *LicensesService) List(opts *ListOptions) (*LicensesResponse, *http.Response, error) {
	return s.ListContext(context.Background(), opts)
}

// ListContext returns a list of licenses with the provided context and
// pagination options.
func (s *LicensesService) ListContext(ctx context.Context, opts *ListOptions) (*LicensesResponse, *http.Response, error) {
	u := "api/v1/licenses"
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

	var licenses LicensesResponse
	resp, err := s.client.Do(req, &licenses)
	if err != nil {
		return nil, resp, err
	}

	return &licenses, resp, nil
}

// Get fetches a single license by its ID.
//
// Unlike the create and update endpoints, this one returns the license object
// itself rather than a {status, payload} envelope, so it returns a *License.
func (s *LicensesService) Get(id int) (*License, *http.Response, error) {
	return s.GetContext(context.Background(), id)
}

// GetContext fetches a single license by its ID with the provided context.
func (s *LicensesService) GetContext(ctx context.Context, id int) (*License, *http.Response, error) {
	req, err := s.client.newRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("api/v1/licenses/%d", id), nil)
	if err != nil {
		return nil, nil, err
	}

	var license License
	resp, err := s.client.Do(req, &license)
	if err != nil {
		return nil, resp, err
	}

	return &license, resp, nil
}

// Create creates a new license.
//
// Name, Seats, and a license-type CategoryID are required by the API.
func (s *LicensesService) Create(license License) (*LicenseResponse, *http.Response, error) {
	return s.CreateContext(context.Background(), license)
}

// CreateContext creates a new license with the provided context.
func (s *LicensesService) CreateContext(ctx context.Context, license License) (*LicenseResponse, *http.Response, error) {
	req, err := s.client.newRequestWithContext(ctx, http.MethodPost, "api/v1/licenses", license)
	if err != nil {
		return nil, nil, err
	}

	var response LicenseResponse
	resp, err := s.client.Do(req, &response)
	if err != nil {
		return nil, resp, err
	}

	return &response, resp, nil
}

// Update updates an existing license, replacing the fields sent.
func (s *LicensesService) Update(id int, license License) (*LicenseResponse, *http.Response, error) {
	return s.UpdateContext(context.Background(), id, license)
}

// UpdateContext updates an existing license with the provided context.
func (s *LicensesService) UpdateContext(ctx context.Context, id int, license License) (*LicenseResponse, *http.Response, error) {
	req, err := s.client.newRequestWithContext(ctx, http.MethodPatch, fmt.Sprintf("api/v1/licenses/%d", id), license)
	if err != nil {
		return nil, nil, err
	}

	var response LicenseResponse
	resp, err := s.client.Do(req, &response)
	if err != nil {
		return nil, resp, err
	}

	return &response, resp, nil
}

// Delete deletes a license.
func (s *LicensesService) Delete(id int) (*Response, *http.Response, error) {
	return s.DeleteContext(context.Background(), id)
}

// DeleteContext deletes a license with the provided context.
func (s *LicensesService) DeleteContext(ctx context.Context, id int) (*Response, *http.Response, error) {
	req, err := s.client.newRequestWithContext(ctx, http.MethodDelete, fmt.Sprintf("api/v1/licenses/%d", id), nil)
	if err != nil {
		return nil, nil, err
	}

	var response Response
	resp, err := s.client.Do(req, &response)
	if err != nil {
		return nil, resp, err
	}

	return &response, resp, nil
}

// ListSeats returns the seats of a license with pagination options.
func (s *LicensesService) ListSeats(licenseID int, opts *ListOptions) (*LicenseSeatsResponse, *http.Response, error) {
	return s.ListSeatsContext(context.Background(), licenseID, opts)
}

// ListSeatsContext returns the seats of a license with the provided context and
// pagination options.
func (s *LicensesService) ListSeatsContext(ctx context.Context, licenseID int, opts *ListOptions) (*LicenseSeatsResponse, *http.Response, error) {
	u := fmt.Sprintf("api/v1/licenses/%d/seats", licenseID)
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

	var seats LicenseSeatsResponse
	resp, err := s.client.Do(req, &seats)
	if err != nil {
		return nil, resp, err
	}

	return &seats, resp, nil
}

// CheckoutSeatToUser checks a license seat out to a user.
func (s *LicensesService) CheckoutSeatToUser(licenseID, seatID, userID int) (*Response, *http.Response, error) {
	return s.CheckoutSeatToUserContext(context.Background(), licenseID, seatID, userID)
}

// CheckoutSeatToUserContext checks a license seat out to a user with the
// provided context.
func (s *LicensesService) CheckoutSeatToUserContext(ctx context.Context, licenseID, seatID, userID int) (*Response, *http.Response, error) {
	return s.patchSeat(ctx, licenseID, seatID, map[string]interface{}{"assigned_to": userID})
}

// CheckoutSeatToAsset checks a license seat out to an asset.
func (s *LicensesService) CheckoutSeatToAsset(licenseID, seatID, assetID int) (*Response, *http.Response, error) {
	return s.CheckoutSeatToAssetContext(context.Background(), licenseID, seatID, assetID)
}

// CheckoutSeatToAssetContext checks a license seat out to an asset with the
// provided context.
func (s *LicensesService) CheckoutSeatToAssetContext(ctx context.Context, licenseID, seatID, assetID int) (*Response, *http.Response, error) {
	return s.patchSeat(ctx, licenseID, seatID, map[string]interface{}{"asset_id": assetID})
}

// CheckinSeat returns a license seat to the pool, clearing both assignments.
func (s *LicensesService) CheckinSeat(licenseID, seatID int) (*Response, *http.Response, error) {
	return s.CheckinSeatContext(context.Background(), licenseID, seatID)
}

// CheckinSeatContext returns a license seat to the pool with the provided
// context.
func (s *LicensesService) CheckinSeatContext(ctx context.Context, licenseID, seatID int) (*Response, *http.Response, error) {
	return s.patchSeat(ctx, licenseID, seatID, map[string]interface{}{"assigned_to": nil, "asset_id": nil})
}

// patchSeat sends an absolute seat assignment. The bodies used by the checkout
// and checkin helpers are idempotent, so a retried request cannot double-apply.
func (s *LicensesService) patchSeat(ctx context.Context, licenseID, seatID int, body map[string]interface{}) (*Response, *http.Response, error) {
	u := fmt.Sprintf("api/v1/licenses/%d/seats/%d", licenseID, seatID)
	req, err := s.client.newRequestWithContext(ctx, http.MethodPatch, u, body)
	if err != nil {
		return nil, nil, err
	}

	var response Response
	resp, err := s.client.Do(req, &response)
	if err != nil {
		return nil, resp, err
	}

	return &response, resp, nil
}
