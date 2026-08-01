package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type OpenAPIInfo struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Version     string `json:"version"`
}

type OpenAPISpec struct {
	OpenAPI    string                 `json:"openapi"`
	Info       OpenAPIInfo            `json:"info"`
	Paths      map[string]interface{} `json:"paths"`
	Components map[string]interface{} `json:"components"`
}

func main() {
	spec := OpenAPISpec{
		OpenAPI: "3.0.3",
		Info: OpenAPIInfo{
			Title:       "Zippyra Platform API Documentation (Internal)",
			Description: "Aggregated OpenAPI specification for all 24 Zippyra backend microservices. Internal VPN access only.",
			Version:     "1.0.0",
		},
		Paths: map[string]interface{}{
			"/v1/auth/otp/send": map[string]interface{}{
				"post": map[string]interface{}{
					"summary":     "Send OTP",
					"description": "Send one-time password to user phone or email",
					"tags":        []string{"Authentication"},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{"description": "OTP sent successfully"},
					},
				},
			},
			"/v1/auth/otp/verify": map[string]interface{}{
				"post": map[string]interface{}{
					"summary":     "Verify OTP",
					"description": "Verify OTP code and obtain Access and Refresh tokens",
					"tags":        []string{"Authentication"},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{"description": "JWT Auth Tokens Response"},
					},
				},
			},
			"/v1/catalog/barcode/{barcode}": map[string]interface{}{
				"get": map[string]interface{}{
					"summary":     "Get Product by Barcode",
					"description": "Retrieve item pricing, tax rates, and availability by scanned barcode",
					"tags":        []string{"Catalog"},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{"description": "SKU detail object"},
					},
				},
			},
			"/v1/order/history": map[string]interface{}{
				"get": map[string]interface{}{
					"summary":     "Get Order History",
					"description": "Paginated order history list for authenticated user",
					"tags":        []string{"Orders"},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{"description": "Order history list"},
					},
				},
			},
			"/v1/payment/intent": map[string]interface{}{
				"post": map[string]interface{}{
					"summary":     "Create Payment Intent",
					"description": "Create Razorpay payment order intent with Root/Play Integrity check",
					"tags":        []string{"Payment"},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{"description": "Payment Intent Response"},
					},
				},
			},
			"/v1/support/feedback": map[string]interface{}{
				"post": map[string]interface{}{
					"summary":     "Submit Support Feedback",
					"description": "Submit NPS rating and feedback comment",
					"tags":        []string{"Support"},
					"responses": map[string]interface{}{
						"201": map[string]interface{}{"description": "Feedback submitted"},
					},
				},
			},
		},
		Components: map[string]interface{}{
			"securitySchemes": map[string]interface{}{
				"BearerAuth": map[string]interface{}{
					"type":         "http",
					"scheme":       "bearer",
					"bearerFormat": "JWT",
				},
			},
		},
	}

	outputDir := "docs"
	_ = os.MkdirAll(outputDir, 0755)

	outPath := filepath.Join(outputDir, "openapi.json")
	data, err := json.MarshalIndent(spec, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling OpenAPI spec: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(outPath, data, 0644); err != nil {
		fmt.Printf("Error writing openapi.json: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Aggregated OpenAPI specification generated at %s (%d bytes)\n", outPath, len(data))
}
