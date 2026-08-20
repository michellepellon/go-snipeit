package snipeit

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestLicensesList(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/licenses" {
			t.Errorf("path = %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"total":1,"rows":[{"id":4,"name":"Google Workspace","seats":25,"free_seats_count":3,"category":{"id":7,"name":"Software"},"reassignable":1}]}`))
	}))
	defer srv.Close()

	c, err := NewClient(srv.URL, "tok")
	if err != nil {
		t.Fatal(err)
	}
	got, _, err := c.Licenses.List(&ListOptions{Limit: 50})
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Rows) != 1 {
		t.Fatalf("rows = %d, want 1", len(got.Rows))
	}
	l := got.Rows[0]
	if l.ID != 4 || l.Name != "Google Workspace" || l.Seats != 25 || l.FreeSeatsCount != 3 {
		t.Errorf("license = %+v", l)
	}
	if l.Category == nil || l.Category.ID != 7 {
		t.Errorf("category = %+v", l.Category)
	}
	if !bool(l.Reassignable) {
		t.Error("reassignable sent as 1 must decode to true")
	}
}

// The API returns nested objects but takes flat *_id fields on write.
func TestLicenseMarshalWritesFlatIDs(t *testing.T) {
	expires := &SnipeTime{Time: time.Date(2027, 3, 1, 0, 0, 0, 0, time.UTC)}
	l := License{
		CommonFields:   CommonFields{Name: "Google Workspace"},
		Seats:          25,
		Category:       &Category{CommonFields: CommonFields{ID: 7}},
		Reassignable:   true,
		PurchaseCost:   "49.99",
		ExpirationDate: expires,
	}
	b, err := json.Marshal(l)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	if m["category_id"] != float64(7) {
		t.Errorf("category_id = %v, want 7 (nested category must flatten)", m["category_id"])
	}
	if m["expiration_date"] != "2027-03-01" {
		t.Errorf("expiration_date = %v, want 2027-03-01", m["expiration_date"])
	}
	if _, ok := m["category"]; ok {
		t.Error("the nested category object must not be sent on write")
	}
	if m["reassignable"] != true || m["seats"] != float64(25) || m["purchase_cost"] != "49.99" {
		t.Errorf("body = %v", m)
	}
}

// A nil date is left alone; a zero date clears the stored value.
func TestLicenseMarshalDateClearing(t *testing.T) {
	b, _ := json.Marshal(License{CommonFields: CommonFields{Name: "L"}})
	var m map[string]interface{}
	_ = json.Unmarshal(b, &m)
	if _, ok := m["expiration_date"]; ok {
		t.Error("a nil ExpirationDate must be omitted, not sent as null")
	}

	b, _ = json.Marshal(License{CommonFields: CommonFields{Name: "L"}, ExpirationDate: &SnipeTime{}})
	_ = json.Unmarshal(b, &m)
	v, ok := m["expiration_date"]
	if !ok || v != nil {
		t.Errorf("a zero ExpirationDate must be sent as null to clear it, got %v (present=%v)", v, ok)
	}
}

func TestLicenseSeatsAndCheckout(t *testing.T) {
	var patched map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/licenses/4/seats":
			_, _ = w.Write([]byte(`{"total":2,"rows":[{"id":11,"license_id":4,"assigned_user":{"id":90}},{"id":12,"license_id":4}]}`))
		case r.Method == http.MethodPatch && r.URL.Path == "/api/v1/licenses/4/seats/12":
			_ = json.NewDecoder(r.Body).Decode(&patched)
			_, _ = w.Write([]byte(`{"status":"success","messages":"Seat updated."}`))
		default:
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
	}))
	defer srv.Close()

	c, err := NewClient(srv.URL, "tok")
	if err != nil {
		t.Fatal(err)
	}
	seats, _, err := c.Licenses.ListSeats(4, &ListOptions{Limit: 500})
	if err != nil {
		t.Fatal(err)
	}
	if len(seats.Rows) != 2 {
		t.Fatalf("seats = %d, want 2", len(seats.Rows))
	}
	if seats.Rows[0].AssignedUser == nil || seats.Rows[0].AssignedUser.ID != 90 {
		t.Errorf("seat 11 assignment = %+v", seats.Rows[0].AssignedUser)
	}
	if seats.Rows[1].AssignedUser != nil || seats.Rows[1].AssignedAsset != nil {
		t.Error("seat 12 must read back as free")
	}

	if _, _, err := c.Licenses.CheckoutSeatToUserContext(context.Background(), 4, 12, 90); err != nil {
		t.Fatal(err)
	}
	if patched["assigned_to"] != float64(90) {
		t.Errorf("checkout body = %v", patched)
	}

	if _, _, err := c.Licenses.CheckinSeat(4, 12); err != nil {
		t.Fatal(err)
	}
	// Check-in must clear both assignment kinds, not just the user.
	if v, ok := patched["assigned_to"]; !ok || v != nil {
		t.Errorf("checkin assigned_to = %v (present=%v), want null", v, ok)
	}
	if v, ok := patched["asset_id"]; !ok || v != nil {
		t.Errorf("checkin asset_id = %v (present=%v), want null", v, ok)
	}
}
