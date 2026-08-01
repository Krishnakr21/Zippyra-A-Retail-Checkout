package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type PostmanInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Schema      string `json:"schema"`
}

type PostmanRequest struct {
	Method string                 `json:"method"`
	Header []map[string]string    `json:"header"`
	Url    map[string]interface{} `json:"url"`
}

type PostmanItem struct {
	Name    string         `json:"name"`
	Request PostmanRequest `json:"request"`
}

type PostmanCollection struct {
	Info PostmanInfo   `json:"info"`
	Item []PostmanItem `json:"item"`
}

func main() {
	collection := PostmanCollection{
		Info: PostmanInfo{
			Name:        "Zippyra Backend Platform APIs",
			Description: "Postman Collection generated from Zippyra OpenAPI specification for developer onboarding and manual testing.",
			Schema:      "https://schema.getpostman.com/json/collection/v2.1.0/collection.json",
		},
		Item: []PostmanItem{
			{
				Name: "Send OTP",
				Request: PostmanRequest{
					Method: "POST",
					Header: []map[string]string{
						{"key": "Content-Type", "value": "application/json"},
					},
					Url: map[string]interface{}{
						"raw":  "{{baseUrl}}/v1/auth/otp/send",
						"host": []string{"{{baseUrl}}"},
						"path": []string{"v1", "auth", "otp", "send"},
					},
				},
			},
			{
				Name: "Verify OTP",
				Request: PostmanRequest{
					Method: "POST",
					Header: []map[string]string{
						{"key": "Content-Type", "value": "application/json"},
					},
					Url: map[string]interface{}{
						"raw":  "{{baseUrl}}/v1/auth/otp/verify",
						"host": []string{"{{baseUrl}}"},
						"path": []string{"v1", "auth", "otp", "verify"},
					},
				},
			},
			{
				Name: "Get Product by Barcode",
				Request: PostmanRequest{
					Method: "GET",
					Header: []map[string]string{},
					Url: map[string]interface{}{
						"raw":  "{{baseUrl}}/v1/catalog/barcode/8901030300011?store_id=store-001",
						"host": []string{"{{baseUrl}}"},
						"path": []string{"v1", "catalog", "barcode", "8901030300011"},
					},
				},
			},
			{
				Name: "Get Order History",
				Request: PostmanRequest{
					Method: "GET",
					Header: []map[string]string{
						{"key": "Authorization", "value": "Bearer {{accessToken}}"},
					},
					Url: map[string]interface{}{
						"raw":  "{{baseUrl}}/v1/order/history",
						"host": []string{"{{baseUrl}}"},
						"path": []string{"v1", "order", "history"},
					},
				},
			},
			{
				Name: "Create Payment Intent",
				Request: PostmanRequest{
					Method: "POST",
					Header: []map[string]string{
						{"key": "Content-Type", "value": "application/json"},
						{"key": "Authorization", "value": "Bearer {{accessToken}}"},
					},
					Url: map[string]interface{}{
						"raw":  "{{baseUrl}}/v1/payment/intent",
						"host": []string{"{{baseUrl}}"},
						"path": []string{"v1", "payment", "intent"},
					},
				},
			},
			{
				Name: "Submit Feedback",
				Request: PostmanRequest{
					Method: "POST",
					Header: []map[string]string{
						{"key": "Content-Type", "value": "application/json"},
						{"key": "Authorization", "value": "Bearer {{accessToken}}"},
					},
					Url: map[string]interface{}{
						"raw":  "{{baseUrl}}/v1/support/feedback",
						"host": []string{"{{baseUrl}}"},
						"path": []string{"v1", "support", "feedback"},
					},
				},
			},
		},
	}

	outputDir := "docs"
	_ = os.MkdirAll(outputDir, 0755)

	outPath := filepath.Join(outputDir, "postman_collection.json")
	data, err := json.MarshalIndent(collection, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling Postman collection: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(outPath, data, 0644); err != nil {
		fmt.Printf("Error writing postman_collection.json: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Postman v2.1 collection exported at %s (%d bytes)\n", outPath, len(data))
}
