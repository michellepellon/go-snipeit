// Package snipeit provides data models for interacting with the Snipe-IT API.
package snipeit

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// FlexInt handles JSON fields that may be returned as either a string or an int.
// The Snipe-IT API is inconsistent about numeric field types — some fields like
// warranty_months may be returned as a quoted string (e.g. "36") instead of a
// bare integer.
type FlexInt int

// UnmarshalJSON implements json.Unmarshaler, accepting both integer and string
// representations of a number.
func (fi *FlexInt) UnmarshalJSON(data []byte) error {
	if string(data) == "null" || string(data) == `""` {
		*fi = 0
		return nil
	}
	// Try int first
	var i int
	if err := json.Unmarshal(data, &i); err == nil {
		*fi = FlexInt(i)
		return nil
	}
	// Try string
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		if s == "" {
			*fi = 0
			return nil
		}
		// Try direct parse first
		if n, err := strconv.Atoi(s); err == nil {
			*fi = FlexInt(n)
			return nil
		}
		// Handle strings like "36 months" — extract leading number
		if idx := strings.IndexFunc(s, func(r rune) bool {
			return r != '-' && (r < '0' || r > '9')
		}); idx > 0 {
			if n, err := strconv.Atoi(strings.TrimSpace(s[:idx])); err == nil {
				*fi = FlexInt(n)
				return nil
			}
		}
		// Non-numeric strings like "None" are treated as zero
		*fi = 0
		return nil
	}
	return fmt.Errorf("FlexInt: cannot unmarshal %s", string(data))
}

// MarshalJSON implements json.Marshaler, always encoding as a bare integer.
func (fi FlexInt) MarshalJSON() ([]byte, error) {
	return json.Marshal(int(fi))
}

// Int returns the underlying int value.
func (fi FlexInt) Int() int {
	return int(fi)
}

// FlexBool handles JSON fields that may be returned as a bool, an int (0/1),
// or a string ("true"/"false"/"0"/"1"). The Snipe-IT API returns some boolean
// fields as integers.
type FlexBool bool

// UnmarshalJSON implements json.Unmarshaler, accepting bool, int (0/1), and
// string ("true"/"false"/"0"/"1") representations.
func (fb *FlexBool) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*fb = false
		return nil
	}
	// Try bool first
	var b bool
	if err := json.Unmarshal(data, &b); err == nil {
		*fb = FlexBool(b)
		return nil
	}
	// Try number (0/1)
	var n int
	if err := json.Unmarshal(data, &n); err == nil {
		*fb = FlexBool(n != 0)
		return nil
	}
	// Try string
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		switch strings.ToLower(s) {
		case "true", "1":
			*fb = true
		default:
			*fb = false
		}
		return nil
	}
	return fmt.Errorf("FlexBool: cannot unmarshal %s", string(data))
}

// MarshalJSON implements json.Marshaler, always encoding as a JSON boolean.
func (fb FlexBool) MarshalJSON() ([]byte, error) {
	return json.Marshal(bool(fb))
}

// Bool returns the underlying bool value.
func (fb FlexBool) Bool() bool {
	return bool(fb)
}

// SnipeTime represents a time field from the Snipe-IT API.
// Snipe-IT returns times as objects with "datetime" and "formatted" fields.
type SnipeTime struct {
	time.Time
}

// UnmarshalJSON implements json.Unmarshaler for SnipeTime.
func (st *SnipeTime) UnmarshalJSON(data []byte) error {
	// Handle null values
	if string(data) == "null" {
		st.Time = time.Time{}
		return nil
	}

	// First try to unmarshal as a simple string (in case API changes or testing)
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		// Try multiple time formats
		formats := []string{
			"2006-01-02 15:04:05",                 // Snipe-IT format
			time.RFC3339,                           // ISO 8601 with timezone
			"2006-01-02T15:04:05.000000Z",        // ISO 8601 with microseconds
			"2006-01-02T15:04:05Z",               // ISO 8601 basic
		}
		
		var parseErr error
		for _, format := range formats {
			t, err := time.Parse(format, str)
			if err == nil {
				st.Time = t
				return nil
			}
			parseErr = err
		}
		return parseErr
	}

	// Otherwise, expect the object format.
	// Snipe-IT uses "datetime" for timestamps and "date" for date-only fields
	// (e.g. purchase_date, warranty_expires).
	var timeObj struct {
		Datetime  string `json:"datetime"`
		Date      string `json:"date"`
		Formatted string `json:"formatted"`
	}
	if err := json.Unmarshal(data, &timeObj); err != nil {
		return err
	}

	if timeObj.Datetime != "" {
		t, err := time.Parse("2006-01-02 15:04:05", timeObj.Datetime)
		if err != nil {
			return err
		}
		st.Time = t
	} else if timeObj.Date != "" {
		t, err := time.Parse("2006-01-02", timeObj.Date)
		if err != nil {
			return err
		}
		st.Time = t
	}

	return nil
}

// MarshalJSON implements json.Marshaler for SnipeTime.
func (st SnipeTime) MarshalJSON() ([]byte, error) {
	if st.Time.IsZero() {
		return []byte("null"), nil
	}
	return json.Marshal(st.Time.Format("2006-01-02 15:04:05"))
}

// FlexUser handles the Snipe-IT API's polymorphic "assigned_to" field.
// On GET responses, the API returns a full User object. On create/update
// response payloads, the API returns just the user's ID as a number, or null.
type FlexUser struct {
	User
}

// UnmarshalJSON implements json.Unmarshaler for FlexUser.
func (fu *FlexUser) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*fu = FlexUser{}
		return nil
	}
	// Try as a number (create/update response)
	var id int
	if err := json.Unmarshal(data, &id); err == nil {
		fu.User = User{}
		fu.User.ID = id
		return nil
	}
	// Otherwise unmarshal as a full User object (GET response)
	var u User
	if err := json.Unmarshal(data, &u); err != nil {
		return fmt.Errorf("FlexUser: cannot unmarshal %s: %w", string(data), err)
	}
	fu.User = u
	return nil
}

// MarshalJSON implements json.Marshaler for FlexUser.
// Always marshals as just the user ID for write operations.
func (fu FlexUser) MarshalJSON() ([]byte, error) {
	if fu.User.ID == 0 {
		return []byte("null"), nil
	}
	return json.Marshal(fu.User.ID)
}

// FlexMessage handles the Snipe-IT API's "messages" field which may be
// returned as a plain string (e.g. "Asset does not exist.") or as an
// object with field-level validation errors (e.g. {"model_id":["The model id field is required."]}).
type FlexMessage string

// UnmarshalJSON implements json.Unmarshaler for FlexMessage.
func (fm *FlexMessage) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*fm = ""
		return nil
	}
	// Try string first
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		*fm = FlexMessage(s)
		return nil
	}
	// Must be an object — store the raw JSON
	*fm = FlexMessage(string(data))
	return nil
}

// String returns the message as a string.
func (fm FlexMessage) String() string {
	return string(fm)
}

// Response represents a standard response structure from the Snipe-IT API.
// Different API endpoints may use different fields within this structure.
// For example, list endpoints typically use Total, Count, and Rows, while
// single-item endpoints typically use Payload.
type Response struct {
	// Status of the API request, typically "success" or "error"
	Status   string      `json:"status"`

	// Message provided by the API, often used for error information.
	// Uses FlexMessage because the API returns this as either a plain string
	// or a JSON object with field-level validation errors.
	Message  FlexMessage `json:"messages,omitempty"`

	// Payload contains the primary data for single-item responses
	Payload  interface{} `json:"payload,omitempty"`

	// Total number of items available (for paginated responses)
	Total    int         `json:"total,omitempty"`

	// Count of items in the current response
	Count    int         `json:"count,omitempty"`

	// Rows contains the data for list/collection responses
	Rows     interface{} `json:"rows,omitempty"`

	// Offset from the beginning of the collection (for pagination)
	Offset   int         `json:"offset,omitempty"`

	// Limit on the number of items returned (for pagination)
	Limit    int         `json:"limit,omitempty"`

	// PageSize indicates the number of items per page (for pagination)
	PageSize int         `json:"pagesize,omitempty"`
}

// CommonFields contains fields that are common across many Snipe-IT resource types.
// This is embedded in other model structs to avoid repetition.
type CommonFields struct {
	// ID is the unique identifier for the resource
	ID          int       `json:"id"`
	
	// CreatedAt is when the resource was created
	CreatedAt   *SnipeTime `json:"created_at"`
	
	// UpdatedAt is when the resource was last updated
	UpdatedAt   *SnipeTime `json:"updated_at"`
	
	// DeletedAt is when the resource was soft-deleted (if applicable)
	DeletedAt   *SnipeTime `json:"deleted_at,omitempty"`
	
	// Name of the resource
	Name        string    `json:"name"`
	
	// Notes associated with the resource
	Notes       string    `json:"notes,omitempty"`
	
	// Available indicates if the resource is available for checkout
	Available   bool      `json:"available"`
	
	// Deleted indicates if the resource has been soft-deleted
	Deleted     bool      `json:"deleted"`
	
	// Image is a URL to the image associated with the resource
	Image       string    `json:"image,omitempty"`
	
	// CustomFields contains any custom fields defined for the resource.
	// When reading from the API, Snipe-IT returns these as a nested object
	// under "custom_fields". When writing, they must be sent as top-level
	// keys (e.g. "_snipeit_ram_2"). The Asset type's MarshalJSON handles
	// this flattening automatically.
	CustomFields map[string]string `json:"-"`
}

// ListOptions specifies common options for paginated API methods.
// These options are used to control pagination, sorting, and filtering of list results.
type ListOptions struct {
	// Page number for paginated results (1-based)
	Page     int `url:"page,omitempty"`
	
	// Limit sets the maximum number of items to return per page
	Limit    int `url:"limit,omitempty"`
	
	// Offset is the number of items to skip before starting to collect results
	Offset   int `url:"offset,omitempty"`
	
	// Sort specifies the field to sort results by (e.g., "id", "name")
	Sort     string `url:"sort,omitempty"`
	
	// SortDir specifies the sort direction, either "asc" or "desc"
	SortDir  string `url:"sort_dir,omitempty"`
	
	// Search is a search term to filter results
	Search   string `url:"search,omitempty"`
}

// Asset represents a Snipe-IT hardware asset.
// Assets are the primary items tracked in Snipe-IT, such as laptops, phones, monitors, etc.
type Asset struct {
	// CommonFields contains standard fields like ID, Name, etc.
	CommonFields
	
	// AssetTag is a unique identifier for the asset in your organization
	AssetTag       string      `json:"asset_tag"`
	
	// Serial is the manufacturer's serial number
	Serial         string      `json:"serial"`
	
	// Model specifies what model the asset is
	Model          Model       `json:"model"`
	
	// ModelNumber is the manufacturer's model number
	ModelNumber    string      `json:"model_number,omitempty"`
	
	// StatusLabel indicates the current status (e.g., "Ready to Deploy", "Deployed")
	StatusLabel    StatusLabel `json:"status_label"`
	
	// Category of the asset (e.g., "Laptop", "Monitor")
	Category       Category    `json:"category"`
	
	// Manufacturer of the asset
	Manufacturer   Manufacturer `json:"manufacturer"`
	
	// Supplier from whom the asset was purchased
	Supplier       Supplier    `json:"supplier,omitempty"`
	
	// Location where the asset is physically located
	Location       Location    `json:"location,omitempty"`
	
	// PurchaseDate when the asset was purchased
	PurchaseDate   *SnipeTime  `json:"purchase_date,omitempty"`
	
	// PurchaseCost of the asset
	PurchaseCost   string      `json:"purchase_cost,omitempty"`
	
	// WarrantyMonths is the length of the warranty in months.
	// Uses FlexInt because the Snipe-IT API may return this as a string.
	WarrantyMonths FlexInt     `json:"warranty_months,omitempty"`
	
	// User to whom the asset is assigned (if any).
	// Uses FlexUser because the API returns a full User object on GET but
	// just a user ID number on create/update response payloads.
	User           *FlexUser   `json:"assigned_to,omitempty"`
	
	// OrderNumber is the order number associated with the asset purchase
	OrderNumber    string      `json:"order_number,omitempty"`

	// AssignedType indicates what type of entity the asset is assigned to
	// (e.g., "user", "location", "asset")
	AssignedType   string      `json:"assigned_type,omitempty"`
}

// UnmarshalJSON implements json.Unmarshaler for Asset.
// The Snipe-IT API returns custom fields as a nested object under
// "custom_fields" where each entry has "field" (db column name) and
// "value". This method extracts those into the CustomFields map keyed
// by db column name, making them easy to compare with write-side values.
func (a *Asset) UnmarshalJSON(data []byte) error {
	// Use an alias to avoid infinite recursion
	type AssetAlias Asset
	var alias AssetAlias
	if err := json.Unmarshal(data, &alias); err != nil {
		return err
	}
	*a = Asset(alias)

	// Parse custom_fields from the raw JSON
	var raw struct {
		CustomFields map[string]struct {
			Field string `json:"field"`
			Value string `json:"value"`
		} `json:"custom_fields"`
	}
	if err := json.Unmarshal(data, &raw); err == nil && len(raw.CustomFields) > 0 {
		a.CustomFields = make(map[string]string, len(raw.CustomFields))
		for _, cf := range raw.CustomFields {
			if cf.Field != "" {
				a.CustomFields[cf.Field] = cf.Value
			}
		}
	}

	return nil
}

// MarshalJSON implements json.Marshaler for Asset.
// The Snipe-IT API returns nested objects for related resources (model,
// status_label, category, etc.) on GET, but expects flat ID fields
// (model_id, status_id, category_id, etc.) on POST/PUT. This method
// converts the nested objects to flat IDs and also flattens custom fields
// to top-level keys.
func (a Asset) MarshalJSON() ([]byte, error) {
	// Build a flat map for the write API
	m := make(map[string]interface{})

	// Core fields
	if a.ID != 0 {
		m["id"] = a.ID
	}
	if a.Name != "" {
		m["name"] = a.Name
	}
	if a.AssetTag != "" {
		m["asset_tag"] = a.AssetTag
	}
	if a.Serial != "" {
		m["serial"] = a.Serial
	}
	if a.ModelNumber != "" {
		m["model_number"] = a.ModelNumber
	}
	if a.Notes != "" {
		m["notes"] = a.Notes
	}
	if a.PurchaseCost != "" {
		m["purchase_cost"] = a.PurchaseCost
	}
	if a.PurchaseDate != nil && !a.PurchaseDate.IsZero() {
		m["purchase_date"] = a.PurchaseDate
	}
	if int(a.WarrantyMonths) != 0 {
		m["warranty_months"] = a.WarrantyMonths
	}
	if a.OrderNumber != "" {
		m["order_number"] = a.OrderNumber
	}
	if a.AssignedType != "" {
		m["assigned_type"] = a.AssignedType
	}
	if a.Image != "" {
		m["image"] = a.Image
	}

	// Flatten nested objects to _id fields for the write API
	if a.Model.ID != 0 {
		m["model_id"] = a.Model.ID
	}
	if a.StatusLabel.ID != 0 {
		m["status_id"] = a.StatusLabel.ID
	}
	if a.Category.ID != 0 {
		m["category_id"] = a.Category.ID
	}
	if a.Manufacturer.ID != 0 {
		m["manufacturer_id"] = a.Manufacturer.ID
	}
	if a.Supplier.ID != 0 {
		m["supplier_id"] = a.Supplier.ID
	}
	if a.Location.ID != 0 {
		m["rtd_location_id"] = a.Location.ID
	}
	if a.User != nil && a.User.User.ID != 0 {
		m["assigned_to"] = a.User.User.ID
	}

	// Flatten custom fields to top-level keys
	for k, v := range a.CustomFields {
		m[k] = v
	}

	return json.Marshal(m)
}

// User represents a Snipe-IT user account.
// Users can check out assets and have assets assigned to them.
type User struct {
	// CommonFields contains standard fields like ID, Name, etc.
	CommonFields
	
	// Username for logging into Snipe-IT
	Username string `json:"username"`
	
	// Email address of the user
	Email    string `json:"email"`
	
	// FirstName of the user
	FirstName string `json:"first_name,omitempty"`
	
	// LastName of the user
	LastName  string `json:"last_name,omitempty"`
	
	// Phone number of the user
	Phone     string `json:"phone,omitempty"`
	
	// JobTitle of the user
	JobTitle  string `json:"jobtitle,omitempty"`
	
	// Employee ID or number
	Employee  string `json:"employee_num,omitempty"`
	
	// Activated indicates if the user account is active
	Activated bool   `json:"activated"`
}

// Model represents a Snipe-IT model.
// Models define a specific type of asset (e.g., "MacBook Pro 16")
// and are associated with Categories and Manufacturers.
type Model struct {
	// CommonFields contains standard fields like ID, Name, etc.
	CommonFields
	
	// ModelNumber is the manufacturer's model identifier
	ModelNumber   string      `json:"model_number,omitempty"`
	
	// Category that this model belongs to
	Category      Category    `json:"category"`
	
	// Manufacturer of this model
	Manufacturer  Manufacturer `json:"manufacturer"`
	
	// FieldsetID is the ID of the custom fieldset associated with this model
	FieldsetID    int         `json:"fieldset_id,omitempty"`
	
	// EOL is the End of Life in months for this model.
	// Uses FlexInt because the Snipe-IT API may return this as a string.
	EOL           FlexInt     `json:"eol,omitempty"`
	
	// AssetsCount is the number of assets of this model
	AssetsCount   int         `json:"assets_count,omitempty"`
}

// MarshalJSON implements json.Marshaler for Model.
// The Snipe-IT API returns nested objects for Category and Manufacturer on GET,
// but expects flat ID fields (category_id, manufacturer_id) on POST/PUT.
func (m Model) MarshalJSON() ([]byte, error) {
	mm := make(map[string]interface{})

	if m.ID != 0 {
		mm["id"] = m.ID
	}
	if m.Name != "" {
		mm["name"] = m.Name
	}
	if m.ModelNumber != "" {
		mm["model_number"] = m.ModelNumber
	}
	if m.Notes != "" {
		mm["notes"] = m.Notes
	}
	if m.FieldsetID != 0 {
		mm["fieldset_id"] = m.FieldsetID
	}
	if m.EOL.Int() != 0 {
		mm["eol"] = m.EOL
	}
	if m.Category.ID != 0 {
		mm["category_id"] = m.Category.ID
	}
	if m.Manufacturer.ID != 0 {
		mm["manufacturer_id"] = m.Manufacturer.ID
	}

	return json.Marshal(mm)
}

// Category represents a Snipe-IT category.
// Categories group models into logical collections (e.g., "Laptops", "Monitors").
type Category struct {
	// CommonFields contains standard fields like ID, Name, etc.
	CommonFields
	
	// Type of category (e.g., "asset", "accessory", "consumable", "component")
	Type          string `json:"type"`
	
	// EULA indicates if this category requires a EULA acceptance
	EULA          bool   `json:"eula,omitempty"`
	
	// Checkin indicates if email should be sent on checkin
	Checkin       bool   `json:"checkin_email,omitempty"`
	
	// Checkout indicates if email should be sent on checkout
	Checkout      bool   `json:"checkout_email,omitempty"`
	
	// RequireMAAC indicates if manager acceptance is required
	RequireMAAC   bool   `json:"require_acceptance,omitempty"`
	
	// AssetsCount is the number of assets in this category
	AssetsCount   int    `json:"assets_count,omitempty"`
	
	// ModelsCount is the number of models in this category
	ModelsCount   int    `json:"models_count,omitempty"`
}

// Manufacturer represents a Snipe-IT manufacturer.
// Manufacturers are companies that make the assets (e.g., "Apple", "Dell").
type Manufacturer struct {
	// CommonFields contains standard fields like ID, Name, etc.
	CommonFields
	
	// URL is the manufacturer's website
	URL          string `json:"url,omitempty"`
	
	// SupportURL is the URL for getting support
	SupportURL   string `json:"support_url,omitempty"`
	
	// SupportPhone is the phone number for getting support
	SupportPhone string `json:"support_phone,omitempty"`
	
	// SupportEmail is the email for getting support
	SupportEmail string `json:"support_email,omitempty"`
	
	// AssetsCount is the number of assets from this manufacturer
	AssetsCount  int    `json:"assets_count,omitempty"`
}

// Location represents a Snipe-IT location.
// Locations are physical places where assets can be assigned or checked out to.
type Location struct {
	// CommonFields contains standard fields like ID, Name, etc.
	CommonFields
	
	// Address line 1
	Address    string     `json:"address,omitempty"`
	
	// Address line 2
	Address2   string     `json:"address2,omitempty"`
	
	// City name
	City       string     `json:"city,omitempty"`
	
	// State or province
	State      string     `json:"state,omitempty"`
	
	// Country name
	Country    string     `json:"country,omitempty"`
	
	// Zip or postal code
	Zip        string     `json:"zip,omitempty"`
	
	// Currency used at this location
	Currency   string     `json:"currency,omitempty"`
	
	// ParentID is the ID of the parent location (for hierarchical locations)
	ParentID   int        `json:"parent_id,omitempty"`
	
	// Parent is the parent location object (for hierarchical locations)
	Parent     *Location  `json:"parent,omitempty"`
	
	// Children are the child locations of this location
	Children   []Location `json:"children,omitempty"`
	
	// AssetsCount is the number of assets at this location
	AssetsCount int       `json:"assets_count,omitempty"`
}

// StatusLabel represents a Snipe-IT status label.
// Status labels define the current state of an asset (e.g., "Ready to Deploy", "Deployed").
type StatusLabel struct {
	// CommonFields contains standard fields like ID, Name, etc.
	CommonFields
	
	// Type of status (typically "deployable", "undeployable" or "archived")
	Type       string `json:"type"`
	
	// StatusMeta provides metadata about the status
	StatusMeta string `json:"status_meta"`
	
	// StatusType indicates the deployment status (typically same as Type)
	StatusType string `json:"status_type"`
}

// Supplier represents a Snipe-IT supplier.
// Suppliers are vendors or companies from whom assets are purchased.
type Supplier struct {
	// CommonFields contains standard fields like ID, Name, etc.
	CommonFields
	
	// Address line 1
	Address    string `json:"address,omitempty"`
	
	// Address line 2
	Address2   string `json:"address2,omitempty"`
	
	// City name
	City       string `json:"city,omitempty"`
	
	// State or province
	State      string `json:"state,omitempty"`
	
	// Country name
	Country    string `json:"country,omitempty"`
	
	// Zip or postal code
	Zip        string `json:"zip,omitempty"`
	
	// ContactName is the name of the primary contact at the supplier
	ContactName string `json:"contact,omitempty"`
	
	// Phone number of the supplier
	Phone      string `json:"phone,omitempty"`
	
	// Fax number of the supplier
	Fax        string `json:"fax,omitempty"`
	
	// Email address for the supplier
	Email      string `json:"email,omitempty"`
	
	// URL is the supplier's website
	URL        string `json:"url,omitempty"`
	
	// AssetsCount is the number of assets from this supplier
	AssetsCount int    `json:"assets_count,omitempty"`
}

// Field represents a Snipe-IT custom field.
// Custom fields extend the data stored on assets beyond the built-in fields.
type Field struct {
	// CommonFields contains standard fields like ID, Name, etc.
	CommonFields

	// DBColumnName is the database column name (e.g. "_snipeit_ram_2")
	DBColumnName string `json:"db_column_name,omitempty"`

	// Element is the form element type (e.g. "text", "textarea", "checkbox")
	Element string `json:"element,omitempty"`

	// Format is the validation format (e.g. "numeric", "BOOLEAN", "URL")
	Format string `json:"format,omitempty"`

	// HelpText is displayed to users when filling in the field
	HelpText string `json:"help_text,omitempty"`

	// FieldValues contains possible values for list-type fields
	FieldValues string `json:"field_values,omitempty"`

	// FieldValuesArray contains possible values as a slice
	FieldValuesArray []string `json:"field_values_array,omitempty"`

	// ShowInListView indicates if the field is shown in asset list views.
	// Uses FlexBool because the Snipe-IT API returns this as 0/1 instead of bool.
	ShowInListView FlexBool `json:"show_in_listview,omitempty"`

	// Type is the field type
	Type string `json:"type,omitempty"`
}

// Fieldset represents a Snipe-IT custom fieldset.
// Fieldsets group custom fields together and are associated with models.
type Fieldset struct {
	// CommonFields contains standard fields like ID, Name, etc.
	CommonFields

	// Fields contains the custom fields in this fieldset as raw JSON.
	// The Snipe-IT API may return this as an array of objects (when fields exist)
	// or as an empty object {} (when no fields are associated), so we use
	// json.RawMessage to handle both cases gracefully.
	Fields json.RawMessage `json:"fields,omitempty"`

	// ModelsCount is the number of models using this fieldset
	ModelsCount int `json:"models_count,omitempty"`
}