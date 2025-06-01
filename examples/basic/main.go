package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/michellepellon/go-snipeit"
)

func main() {
	// Get Snipe-IT URL and API token from environment variables
	snipeURL := os.Getenv("SNIPEIT_URL")
	apiToken := os.Getenv("SNIPEIT_API_TOKEN")

	if snipeURL == "" || apiToken == "" {
		log.Fatal("SNIPEIT_URL and SNIPEIT_API_TOKEN environment variables must be set")
	}

	// Create a new client with rate limiting and retry capabilities
	client, err := snipeit.NewClientWithOptions(
		snipeURL,
		apiToken,
		&snipeit.ClientOptions{
			// Enable rate limiting - 5 requests per second with burst of 10
			RateLimiter: snipeit.NewTokenBucketRateLimiter(5, 10),
			
			// Custom retry policy
			RetryPolicy: &snipeit.RetryPolicy{
				MaxRetries:          3,
				RetryableStatusCodes: map[int]bool{
					429: true, // Too Many Requests
					500: true, // Internal Server Error
					502: true, // Bad Gateway
					503: true, // Service Unavailable
					504: true, // Gateway Timeout
				},
				InitialBackoff:    500 * time.Millisecond,
				MaxBackoff:        10 * time.Second,
				BackoffMultiplier: 2.0,
				Jitter:            0.2,
			},
		},
	)
	if err != nil {
		log.Fatalf("Error creating client: %v", err)
	}
	
	// Example of using a basic client without rate limiting or retries:
	// client, err := snipeit.NewClient(snipeURL, apiToken)
	// if err != nil {
	//     log.Fatalf("Error creating client: %v", err)
	// }
	
	// Example of using a custom HTTP client with timeout:
	// httpClient := &http.Client{
	//     Timeout: 30 * time.Second,
	// }
	// client, err := snipeit.NewClientWithHTTPClient(httpClient, snipeURL, apiToken)
	// if err != nil {
	//     log.Fatalf("Error creating client: %v", err)
	// }

	// List assets with pagination
	fmt.Println("Listing assets...")
	opts := &snipeit.ListOptions{
		Limit: 5,
		Page:  1,
	}
	assets, _, err := client.Assets.List(opts)
	if err != nil {
		log.Fatalf("Error listing assets: %v", err)
	}

	// Print asset information
	fmt.Printf("Total assets: %d\n", assets.Total)
	fmt.Printf("Number of assets returned: %d\n", assets.Count)
	for _, asset := range assets.Rows {
		fmt.Printf("- ID: %d, Name: %s, Tag: %s\n", asset.ID, asset.Name, asset.AssetTag)
	}

	// If there are assets, show detailed information for the first one
	if len(assets.Rows) > 0 {
		assetID := assets.Rows[0].ID
		fmt.Printf("\nGetting detailed information for asset ID %d...\n", assetID)
		
		// Create a context with a timeout
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		
		// Use the context-aware version of Get
		asset, _, err := client.Assets.GetContext(ctx, assetID)
		if err != nil {
			log.Fatalf("Error getting asset details: %v", err)
		}
		
		fmt.Printf("Asset details:\n")
		fmt.Printf("  Name: %s\n", asset.Name)
		fmt.Printf("  Tag: %s\n", asset.AssetTag)
		fmt.Printf("  Serial: %s\n", asset.Serial)
		fmt.Printf("  Model: %s\n", asset.Model.Name)
		fmt.Printf("  Category: %s\n", asset.Category.Name)
		fmt.Printf("  Status: %s\n", asset.StatusLabel.Name)
		
		if asset.User != nil {
			fmt.Printf("  Assigned to: %s\n", asset.User.Name)
		} else {
			fmt.Printf("  Assigned to: Not assigned\n")
		}
	}
	
	// Demonstrate searching for an asset by serial number
	fmt.Printf("\nSearching for asset by serial number...\n")
	// Note: Replace "DWDFN73" with an actual serial number from your Snipe-IT instance
	serialToSearch := "DWDFN73"
	assetsBySerial, resp, err := client.Assets.GetAssetBySerial(serialToSearch)
	if err != nil {
		// Check if it's a 404 not found error
		if resp != nil && resp.StatusCode == 404 {
			fmt.Printf("Asset with serial number %s not found\n", serialToSearch)
		} else {
			fmt.Printf("Error searching for asset by serial: %v\n", err)
		}
	} else if assetsBySerial.Total == 0 {
		fmt.Printf("No assets found with serial number %s\n", serialToSearch)
	} else {
		fmt.Printf("Found %d asset(s) with serial number %s:\n", assetsBySerial.Total, serialToSearch)
		for _, asset := range assetsBySerial.Rows {
			fmt.Printf("  Name: %s\n", asset.Name)
			fmt.Printf("  Tag: %s\n", asset.AssetTag)
			fmt.Printf("  Serial: %s\n", asset.Serial)
			fmt.Printf("  Model: %s\n", asset.Model.Name)
		}
	}
	
	// Demonstrate how to make concurrent API requests with rate limiting
	fmt.Printf("\nMaking multiple concurrent API requests (rate limited)...\n")
	makeConcurrentRequests(client, 20)
}

// makeConcurrentRequests demonstrates how the rate limiter handles concurrent requests
func makeConcurrentRequests(client *snipeit.Client, count int) {
	start := time.Now()
	
	// Create a wait group to wait for all goroutines to finish
	done := make(chan bool)
	results := make(chan string, count)
	
	// Launch multiple goroutines to make concurrent requests
	for i := 0; i < count; i++ {
		go func(idx int) {
			// Each goroutine makes a request
			opts := &snipeit.ListOptions{
				Limit: 1,
				Page:  1,
			}
			assets, _, err := client.Assets.List(opts)
			if err != nil {
				results <- fmt.Sprintf("Request %d error: %v", idx, err)
			} else {
				results <- fmt.Sprintf("Request %d completed, got %d assets", idx, assets.Count)
			}
		}(i)
	}
	
	// Collect results
	go func() {
		for i := 0; i < count; i++ {
			fmt.Println(<-results)
		}
		done <- true
	}()
	
	// Wait for all requests to complete
	<-done
	
	elapsed := time.Since(start)
	fmt.Printf("\nAll %d requests completed in %v\n", count, elapsed)
	fmt.Printf("Average time per request: %v\n", elapsed/time.Duration(count))
}