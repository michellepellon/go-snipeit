// Package snipeit provides data models for interacting with the Snipe-IT API.
package snipeit

import "time"

// Response represents a standard response structure from the Snipe-IT API.
// Different API endpoints may use different fields within this structure.
// For example, list endpoints typically use Total, Count, and Rows, while
// single-item endpoints typically use Payload.
type Response struct {
	// Status of the API request, typically "success" or "error"
	Status   string      `json:"status"`
	
	// Message provided by the API, often used for error information
	Message  string      `json:"messages,omitempty"`
	
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
	CreatedAt   time.Time `json:"created_at"`
	
	// UpdatedAt is when the resource was last updated
	UpdatedAt   time.Time `json:"updated_at"`
	
	// DeletedAt is when the resource was soft-deleted (if applicable)
	DeletedAt   time.Time `json:"deleted_at,omitempty"`
	
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
	
	// CustomFields contains any custom fields defined for the resource
	CustomFields struct{} `json:"custom_fields,omitempty"`
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
	PurchaseDate   *time.Time  `json:"purchase_date,omitempty"`
	
	// PurchaseCost of the asset
	PurchaseCost   string      `json:"purchase_cost,omitempty"`
	
	// WarrantyMonths is the length of the warranty in months
	WarrantyMonths int         `json:"warranty_months,omitempty"`
	
	// User to whom the asset is assigned (if any)
	User           *User       `json:"assigned_to,omitempty"`
	
	// AssignedType indicates what type of entity the asset is assigned to
	// (e.g., "user", "location", "asset")
	AssignedType   string      `json:"assigned_type,omitempty"`
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
	
	// EOL is the End of Life in months for this model
	EOL           int         `json:"eol,omitempty"`
	
	// AssetsCount is the number of assets of this model
	AssetsCount   int         `json:"assets_count,omitempty"`
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