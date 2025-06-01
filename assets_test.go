package snipeit

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"testing"
	"time"
)

func TestAssetsList(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	mux.HandleFunc("/api/v1/hardware", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		testHeader(t, r, "Accept", "application/json")
		testHeader(t, r, "Authorization", "Bearer test-token")

		// Check query parameters
		if r.URL.Query().Get("limit") != "5" {
			t.Errorf("Request URL query parameter 'limit' = %v, expected %v", r.URL.Query().Get("limit"), "5")
		}
		if r.URL.Query().Get("offset") != "10" {
			t.Errorf("Request URL query parameter 'offset' = %v, expected %v", r.URL.Query().Get("offset"), "10")
		}

		fmt.Fprint(w, `{
			"status": "success",
			"total": 2,
			"count": 2,
			"rows": [
				{
					"id": 1,
					"name": "Asset 1",
					"asset_tag": "AT-1",
					"serial": "SN-1",
					"model": {
						"id": 1,
						"name": "Model 1"
					},
					"status_label": {
						"id": 1,
						"name": "Ready to Deploy",
						"status_type": "deployable",
						"status_meta": "deployable"
					},
					"category": {
						"id": 1,
						"name": "Laptop"
					},
					"manufacturer": {
						"id": 1,
						"name": "Manufacturer 1"
					},
					"created_at": "2023-01-01T12:00:00.000000Z",
					"updated_at": "2023-01-01T12:00:00.000000Z",
					"available": true,
					"deleted": false
				},
				{
					"id": 2,
					"name": "Asset 2",
					"asset_tag": "AT-2",
					"serial": "SN-2",
					"model": {
						"id": 2,
						"name": "Model 2"
					},
					"status_label": {
						"id": 2,
						"name": "Deployed",
						"status_type": "deployed",
						"status_meta": "deployed"
					},
					"category": {
						"id": 1,
						"name": "Laptop"
					},
					"manufacturer": {
						"id": 1,
						"name": "Manufacturer 1"
					},
					"created_at": "2023-01-02T12:00:00.000000Z",
					"updated_at": "2023-01-02T12:00:00.000000Z",
					"available": false,
					"deleted": false
				}
			]
		}`)
	})

	opts := &ListOptions{
		Limit:  5,
		Offset: 10,
	}
	
	assets, _, err := client.Assets.List(opts)
	if err != nil {
		t.Fatalf("Assets.List returned error: %v", err)
	}

	if assets.Total != 2 {
		t.Errorf("Assets.List returned Total = %d, expected %d", assets.Total, 2)
	}
	
	if assets.Count != 2 {
		t.Errorf("Assets.List returned Count = %d, expected %d", assets.Count, 2)
	}
	
	if len(assets.Rows) != 2 {
		t.Errorf("Assets.List returned %d assets, expected %d", len(assets.Rows), 2)
	}

	// Check the first asset
	createdAt, _ := time.Parse(time.RFC3339, "2023-01-01T12:00:00.000000Z")
	updatedAt, _ := time.Parse(time.RFC3339, "2023-01-01T12:00:00.000000Z")
	
	// Test context timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	
	// Sleep to ensure the context times out
	time.Sleep(20 * time.Millisecond)
	
	// Execute the request with a timed-out context
	_, _, err = client.Assets.ListContext(ctx, opts)
	if err == nil {
		t.Error("Expected context timeout error, got nil")
	} else if err != context.DeadlineExceeded {
		t.Errorf("Expected context.DeadlineExceeded error, got %v", err)
	}
	
	expectedAsset1 := Asset{
		CommonFields: CommonFields{
			ID:        1,
			Name:      "Asset 1",
			CreatedAt: &SnipeTime{createdAt},
			UpdatedAt: &SnipeTime{updatedAt},
			Available: true,
			Deleted:   false,
		},
		AssetTag: "AT-1",
		Serial:   "SN-1",
		Model: Model{
			CommonFields: CommonFields{
				ID:   1,
				Name: "Model 1",
			},
		},
		StatusLabel: StatusLabel{
			CommonFields: CommonFields{
				ID:   1,
				Name: "Ready to Deploy",
			},
			StatusType: "deployable",
			StatusMeta: "deployable",
		},
		Category: Category{
			CommonFields: CommonFields{
				ID:   1,
				Name: "Laptop",
			},
		},
		Manufacturer: Manufacturer{
			CommonFields: CommonFields{
				ID:   1,
				Name: "Manufacturer 1",
			},
		},
	}

	if !reflect.DeepEqual(assets.Rows[0], expectedAsset1) {
		t.Errorf("Assets.List first asset = %+v, expected %+v", assets.Rows[0], expectedAsset1)
	}
}

func TestAssetsGet(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	mux.HandleFunc("/api/v1/hardware/1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"status": "success",
			"payload": {
				"id": 1,
				"name": "Asset 1",
				"asset_tag": "AT-1",
				"serial": "SN-1",
				"model": {
					"id": 1,
					"name": "Model 1"
				},
				"status_label": {
					"id": 1,
					"name": "Ready to Deploy",
					"status_type": "deployable",
					"status_meta": "deployable"
				},
				"category": {
					"id": 1,
					"name": "Laptop"
				},
				"manufacturer": {
					"id": 1,
					"name": "Manufacturer 1"
				},
				"created_at": "2023-01-01T12:00:00.000000Z",
				"updated_at": "2023-01-01T12:00:00.000000Z",
				"available": true,
				"deleted": false
			}
		}`)
	})

	asset, _, err := client.Assets.Get(1)
	if err != nil {
		t.Fatalf("Assets.Get returned error: %v", err)
	}

	if asset.Status != "success" {
		t.Errorf("Assets.Get returned Status = %s, expected %s", asset.Status, "success")
	}

	createdAt, _ := time.Parse(time.RFC3339, "2023-01-01T12:00:00.000000Z")
	updatedAt, _ := time.Parse(time.RFC3339, "2023-01-01T12:00:00.000000Z")
	
	expectedAsset := Asset{
		CommonFields: CommonFields{
			ID:        1,
			Name:      "Asset 1",
			CreatedAt: &SnipeTime{createdAt},
			UpdatedAt: &SnipeTime{updatedAt},
			Available: true,
			Deleted:   false,
		},
		AssetTag: "AT-1",
		Serial:   "SN-1",
		Model: Model{
			CommonFields: CommonFields{
				ID:   1,
				Name: "Model 1",
			},
		},
		StatusLabel: StatusLabel{
			CommonFields: CommonFields{
				ID:   1,
				Name: "Ready to Deploy",
			},
			StatusType: "deployable",
			StatusMeta: "deployable",
		},
		Category: Category{
			CommonFields: CommonFields{
				ID:   1,
				Name: "Laptop",
			},
		},
		Manufacturer: Manufacturer{
			CommonFields: CommonFields{
				ID:   1,
				Name: "Manufacturer 1",
			},
		},
	}

	if !reflect.DeepEqual(asset.Asset, expectedAsset) {
		t.Errorf("Assets.Get returned = %+v, expected %+v", asset.Asset, expectedAsset)
	}
}

func TestAssetsCreate(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	mux.HandleFunc("/api/v1/hardware", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		testHeader(t, r, "Content-Type", "application/json")
		
		// Verify the asset was serialized correctly
		var requestBody map[string]interface{}
		json.NewDecoder(r.Body).Decode(&requestBody)
		
		if requestBody["asset_tag"] != "NEW-1" {
			t.Errorf("Request body asset_tag = %v, expected %v", requestBody["asset_tag"], "NEW-1")
		}

		fmt.Fprint(w, `{
			"status": "success",
			"payload": {
				"id": 3,
				"name": "New Asset",
				"asset_tag": "NEW-1",
				"serial": "SN-NEW-1",
				"model": {
					"id": 1,
					"name": "Model 1"
				},
				"status_label": {
					"id": 1,
					"name": "Ready to Deploy",
					"status_type": "deployable",
					"status_meta": "deployable"
				},
				"category": {
					"id": 1,
					"name": "Laptop"
				},
				"manufacturer": {
					"id": 1,
					"name": "Manufacturer 1"
				},
				"created_at": "2023-01-03T12:00:00.000000Z",
				"updated_at": "2023-01-03T12:00:00.000000Z",
				"available": true,
				"deleted": false
			}
		}`)
	})

	newAsset := Asset{
		AssetTag: "NEW-1",
		Serial:   "SN-NEW-1",
		CommonFields: CommonFields{
			Name: "New Asset",
		},
		Model: Model{
			CommonFields: CommonFields{
				ID: 1,
			},
		},
		StatusLabel: StatusLabel{
			CommonFields: CommonFields{
				ID: 1,
			},
		},
	}

	asset, _, err := client.Assets.Create(newAsset)
	if err != nil {
		t.Fatalf("Assets.Create returned error: %v", err)
	}

	if asset.Status != "success" {
		t.Errorf("Assets.Create returned Status = %s, expected %s", asset.Status, "success")
	}

	if asset.Payload.ID != 3 {
		t.Errorf("Assets.Create returned ID = %d, expected %d", asset.Payload.ID, 3)
	}
}

func TestAssetsUpdate(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	mux.HandleFunc("/api/v1/hardware/1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		
		// Verify the asset was serialized correctly
		var requestBody map[string]interface{}
		json.NewDecoder(r.Body).Decode(&requestBody)
		
		if requestBody["name"] != "Updated Asset" {
			t.Errorf("Request body name = %v, expected %v", requestBody["name"], "Updated Asset")
		}

		fmt.Fprint(w, `{
			"status": "success",
			"payload": {
				"id": 1,
				"name": "Updated Asset",
				"asset_tag": "AT-1",
				"serial": "SN-1",
				"model": {
					"id": 1,
					"name": "Model 1"
				},
				"status_label": {
					"id": 1,
					"name": "Ready to Deploy",
					"status_type": "deployable",
					"status_meta": "deployable"
				},
				"category": {
					"id": 1,
					"name": "Laptop"
				},
				"manufacturer": {
					"id": 1,
					"name": "Manufacturer 1"
				},
				"created_at": "2023-01-01T12:00:00.000000Z",
				"updated_at": "2023-01-03T12:00:00.000000Z",
				"available": true,
				"deleted": false
			}
		}`)
	})

	updateAsset := Asset{
		CommonFields: CommonFields{
			Name: "Updated Asset",
		},
	}

	asset, _, err := client.Assets.Update(1, updateAsset)
	if err != nil {
		t.Fatalf("Assets.Update returned error: %v", err)
	}

	if asset.Status != "success" {
		t.Errorf("Assets.Update returned Status = %s, expected %s", asset.Status, "success")
	}

	if asset.Payload.Name != "Updated Asset" {
		t.Errorf("Assets.Update returned Name = %s, expected %s", asset.Payload.Name, "Updated Asset")
	}
}

func TestAssetsDelete(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	mux.HandleFunc("/api/v1/hardware/1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{
			"status": "success",
			"message": "Asset deleted successfully"
		}`)
	})

	resp, err := client.Assets.Delete(1)
	if err != nil {
		t.Fatalf("Assets.Delete returned error: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Assets.Delete returned status code = %d, expected %d", resp.StatusCode, http.StatusOK)
	}
}

func TestAssetsCheckout(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	mux.HandleFunc("/api/v1/hardware/1/checkout", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		
		// Verify request body
		var requestBody map[string]interface{}
		json.NewDecoder(r.Body).Decode(&requestBody)
		
		if requestBody["assigned_user"] != float64(2) {
			t.Errorf("Request body assigned_user = %v, expected %v", requestBody["assigned_user"], float64(2))
		}

		fmt.Fprint(w, `{
			"status": "success",
			"payload": {
				"id": 1,
				"name": "Asset 1",
				"asset_tag": "AT-1",
				"serial": "SN-1",
				"assigned_to": {
					"id": 2,
					"name": "John Doe",
					"username": "johndoe"
				},
				"assigned_type": "user",
				"available": false,
				"deleted": false
			}
		}`)
	})

	checkout := map[string]interface{}{
		"assigned_user": 2,
		"checkout_at":   "2023-01-03",
		"note":          "Assigned to John Doe",
	}

	asset, _, err := client.Assets.Checkout(1, checkout)
	if err != nil {
		t.Fatalf("Assets.Checkout returned error: %v", err)
	}

	if asset.Status != "success" {
		t.Errorf("Assets.Checkout returned Status = %s, expected %s", asset.Status, "success")
	}

	if asset.Payload.User == nil || asset.Payload.User.ID != 2 {
		t.Errorf("Assets.Checkout assigned_to user ID = %v, expected %v", 
			asset.Payload.User, 2)
	}

	if asset.Payload.AssignedType != "user" {
		t.Errorf("Assets.Checkout assigned_type = %s, expected %s", 
			asset.Payload.AssignedType, "user")
	}

	if asset.Payload.Available != false {
		t.Errorf("Assets.Checkout available = %v, expected %v", 
			asset.Payload.Available, false)
	}
}

func TestAssetsCheckin(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	mux.HandleFunc("/api/v1/hardware/1/checkin", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		
		// Verify request body
		var requestBody map[string]interface{}
		json.NewDecoder(r.Body).Decode(&requestBody)
		
		if requestBody["note"] != "Checked in from John Doe" {
			t.Errorf("Request body note = %v, expected %v", 
				requestBody["note"], "Checked in from John Doe")
		}

		fmt.Fprint(w, `{
			"status": "success",
			"payload": {
				"id": 1,
				"name": "Asset 1",
				"asset_tag": "AT-1",
				"serial": "SN-1",
				"assigned_to": null,
				"assigned_type": null,
				"available": true,
				"deleted": false
			}
		}`)
	})

	checkin := map[string]interface{}{
		"note": "Checked in from John Doe",
	}

	asset, _, err := client.Assets.Checkin(1, checkin)
	if err != nil {
		t.Fatalf("Assets.Checkin returned error: %v", err)
	}

	if asset.Status != "success" {
		t.Errorf("Assets.Checkin returned Status = %s, expected %s", asset.Status, "success")
	}

	if asset.User != nil {
		t.Errorf("Assets.Checkin assigned_to = %v, expected %v", 
			asset.User, nil)
	}

	if asset.Available != true {
		t.Errorf("Assets.Checkin available = %v, expected %v", 
			asset.Available, true)
	}
}

func TestAssetsGetAssetBySerial(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	mux.HandleFunc("/api/v1/hardware/byserial/SN-12345", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		testHeader(t, r, "Accept", "application/json")
		testHeader(t, r, "Authorization", "Bearer test-token")
		fmt.Fprint(w, `{
			"total": 1,
			"rows": [{
				"id": 1,
				"name": "Asset 1",
				"asset_tag": "AT-1",
				"serial": "SN-12345",
				"model": {
					"id": 1,
					"name": "Model 1"
				},
				"status_label": {
					"id": 1,
					"name": "Ready to Deploy",
					"status_type": "deployable",
					"status_meta": "deployable"
				},
				"category": {
					"id": 1,
					"name": "Laptop"
				},
				"manufacturer": {
					"id": 1,
					"name": "Manufacturer 1"
				},
				"created_at": "2023-01-01T12:00:00.000000Z",
				"updated_at": "2023-01-01T12:00:00.000000Z",
				"available": true,
				"deleted": false
			}]
		}`)
	})

	asset, _, err := client.Assets.GetAssetBySerial("SN-12345")
	if err != nil {
		t.Fatalf("Assets.GetAssetBySerial returned error: %v", err)
	}

	if asset.Status != "success" {
		t.Errorf("Assets.GetAssetBySerial returned Status = %s, expected %s", asset.Status, "success")
	}

	createdAt, _ := time.Parse(time.RFC3339, "2023-01-01T12:00:00.000000Z")
	updatedAt, _ := time.Parse(time.RFC3339, "2023-01-01T12:00:00.000000Z")
	
	expectedAsset := Asset{
		CommonFields: CommonFields{
			ID:        1,
			Name:      "Asset 1",
			CreatedAt: &SnipeTime{createdAt},
			UpdatedAt: &SnipeTime{updatedAt},
			Available: true,
			Deleted:   false,
		},
		AssetTag: "AT-1",
		Serial:   "SN-12345",
		Model: Model{
			CommonFields: CommonFields{
				ID:   1,
				Name: "Model 1",
			},
		},
		StatusLabel: StatusLabel{
			CommonFields: CommonFields{
				ID:   1,
				Name: "Ready to Deploy",
			},
			StatusType: "deployable",
			StatusMeta: "deployable",
		},
		Category: Category{
			CommonFields: CommonFields{
				ID:   1,
				Name: "Laptop",
			},
		},
		Manufacturer: Manufacturer{
			CommonFields: CommonFields{
				ID:   1,
				Name: "Manufacturer 1",
			},
		},
	}

	if !reflect.DeepEqual(asset, expectedAsset) {
		t.Errorf("Assets.GetAssetBySerial returned = %+v, expected %+v", asset, expectedAsset)
	}
}

func TestAssetsGetAssetBySerialNotFound(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	mux.HandleFunc("/api/v1/hardware/byserial/INVALID-SERIAL", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{
			"status": "error",
			"message": "Asset not found."
		}`)
	})

	_, resp, err := client.Assets.GetAssetBySerial("INVALID-SERIAL")
	if err == nil {
		t.Fatal("Assets.GetAssetBySerial expected error for not found, got none")
	}

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Assets.GetAssetBySerial returned status code = %d, expected %d", resp.StatusCode, http.StatusNotFound)
	}

	errorResponse, ok := err.(*ErrorResponse)
	if !ok {
		t.Fatalf("Assets.GetAssetBySerial error type = %T, expected *ErrorResponse", err)
	}

	if errorResponse.Message != "Asset not found." {
		t.Errorf("ErrorResponse.Message = %q, expected %q", errorResponse.Message, "Asset not found.")
	}
}

func TestAssetsGetAssetBySerialContext(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	mux.HandleFunc("/api/v1/hardware/byserial/SN-CONTEXT", func(w http.ResponseWriter, r *http.Request) {
		// Simulate a delay to test context timeout
		time.Sleep(50 * time.Millisecond)
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"status": "success",
			"payload": {
				"id": 1,
				"name": "Asset 1",
				"asset_tag": "AT-1",
				"serial": "SN-CONTEXT"
			}
		}`)
	})

	// Test context timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, _, err := client.Assets.GetAssetBySerialContext(ctx, "SN-CONTEXT")
	if err == nil {
		t.Error("Expected context timeout error, got nil")
	} else if err != context.DeadlineExceeded {
		t.Errorf("Expected context.DeadlineExceeded error, got %v", err)
	}
}