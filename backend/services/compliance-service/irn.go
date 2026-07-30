package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type IRPClient interface {
	SubmitEInvoice(ctx context.Context, payload map[string]interface{}) (*IRPSubmitResponse, error)
}

type IRPSubmitResponse struct {
	Status       string `json:"status"` // "SUCCESS" | "FAILED"
	IRN          string `json:"irn"`
	AckNo        string `json:"ack_no"`
	AckDate      string `json:"ack_date"`
	SignedQRCode string `json:"signed_qr_code"`
	ErrorDetails string `json:"error_details"`
}

type HTTPIRPClient struct {
	baseURL    string
	username   string
	password   string
	clientID   string
	httpClient *http.Client
}

func NewHTTPIRPClient(baseURL, username, password, clientID string) *HTTPIRPClient {
	return &HTTPIRPClient{
		baseURL:  baseURL,
		username: username,
		password: password,
		clientID: clientID,
		httpClient: &http.Client{
			Timeout: 15 * time.Second, // 15s timeout for government IRP portal
		},
	}
}

func (c *HTTPIRPClient) SubmitEInvoice(ctx context.Context, payload map[string]interface{}) (*IRPSubmitResponse, error) {
	if c.baseURL == "" {
		// Mock response for dev/test environment
		orderID := "mock"
		if docDtls, ok := payload["DocDtls"].(map[string]interface{}); ok {
			if num, ok := docDtls["No"].(string); ok {
				orderID = num
			}
		}

		nowStr := time.Now().Format("2006-01-02 15:04:05")
		return &IRPSubmitResponse{
			Status:       "SUCCESS",
			IRN:          fmt.Sprintf("irn_64char_hash_mock_%s_%d", orderID, time.Now().Unix()),
			AckNo:        fmt.Sprintf("12260731%04d", time.Now().Unix()%10000),
			AckDate:      nowStr,
			SignedQRCode: "data:image/png;base64,mock_signed_irp_qr_code",
		}, nil
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal IRP payload: %w", err)
	}

	url := fmt.Sprintf("%s/eInvoice/Generate", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create IRP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("user_name", c.username)
	req.Header.Set("password", c.password)
	req.Header.Set("client_id", c.clientID)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("IRP HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read IRP response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("IRP returned non-200 status %d: %s", resp.StatusCode, string(respBytes))
	}

	var irpResp IRPSubmitResponse
	if err := json.Unmarshal(respBytes, &irpResp); err != nil {
		return nil, fmt.Errorf("failed to parse IRP response: %w", err)
	}

	return &irpResp, nil
}

// BuildCanonicalIRPPayload builds standard GST E-Invoice payload
func BuildCanonicalIRPPayload(orderID, storeID, storeGSTIN string, items []map[string]interface{}, totalPaise, cgstPaise, sgstPaise, igstPaise int64) map[string]interface{} {
	now := time.Now()
	docNo := orderID
	if len(docNo) > 16 {
		docNo = docNo[len(docNo)-16:]
	}

	var itemList []map[string]interface{}
	for idx, item := range items {
		hsn := "8471"
		if h, ok := item["hsn_code"].(string); ok && h != "" {
			hsn = h
		}
		name := "Retail Item"
		if n, ok := item["name"].(string); ok && n != "" {
			name = n
		}
		qty := 1
		if q, ok := item["qty"].(float64); ok {
			qty = int(q)
		} else if q, ok := item["qty"].(int); ok {
			qty = q
		}
		pricePaise := int64(0)
		if p, ok := item["price_paise"].(float64); ok {
			pricePaise = int64(p)
		}

		itemValRupees := float64(pricePaise*int64(qty)) / 100.0

		itemList = append(itemList, map[string]interface{}{
			"SlNo":        fmt.Sprintf("%d", idx+1),
			"PrdDesc":     name,
			"IsServc":     "N",
			"HsnCd":       hsn,
			"Qty":         qty,
			"UnitPrice":   float64(pricePaise) / 100.0,
			"TotAmt":      itemValRupees,
			"AssAmt":      itemValRupees,
			"GstRt":       18.0,
			"TotItemVal":  itemValRupees,
		})
	}

	totValRupees := float64(totalPaise) / 100.0
	cgstRupees := float64(cgstPaise) / 100.0
	sgstRupees := float64(sgstPaise) / 100.0
	igstRupees := float64(igstPaise) / 100.0

	return map[string]interface{}{
		"Version": "1.03",
		"TranDtls": map[string]interface{}{
			"TaxSch": "GST",
			"SupTyp": "B2C",
		},
		"DocDtls": map[string]interface{}{
			"Typ": "INV",
			"No":  docNo,
			"Dt":  now.Format("02/01/2006"),
		},
		"SellerDtls": map[string]interface{}{
			"Gstin": storeGSTIN,
			"LglNm": "Zippyra Store",
			"Addr1": "Store Location",
			"Loc":   "Bengaluru",
			"Pin":   560001,
			"Stcd":  "29",
		},
		"BuyerDtls": map[string]interface{}{
			"Gstin": "URP",
			"LglNm": "Walk-in Retail Customer",
			"Pos":   "29",
		},
		"ItemList": itemList,
		"ValDtls": map[string]interface{}{
			"AssVal":   totValRupees - (cgstRupees + sgstRupees + igstRupees),
			"CgstVal":  cgstRupees,
			"SgstVal":  sgstRupees,
			"IgstVal":  igstRupees,
			"TotInvVal": totValRupees,
		},
	}
}
