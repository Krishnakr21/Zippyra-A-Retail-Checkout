package tally_adapter

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/zippyra/zippyra-connector/internal/erp_adapter"
	"github.com/zippyra/zippyra-connector/internal/logging"
)

type Adapter struct {
	endpoint   string
	httpClient *http.Client
	logger     *logging.Logger
}

func NewAdapter(endpoint string, logger *logging.Logger) *Adapter {
	if endpoint == "" {
		endpoint = "http://127.0.0.1:9000"
	}
	return &Adapter{
		endpoint: strings.TrimRight(endpoint, "/"),
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger: logger,
	}
}

func (a *Adapter) HealthCheck(ctx context.Context) error {
	xmlReq := `<ENVELOPE>
		<HEADER>
			<TALLYREQUEST>Export Data</TALLYREQUEST>
		</HEADER>
		<BODY>
			<EXPORTDATA>
				<REQUESTDESC>
					<REPORTNAME>List of Companies</REPORTNAME>
				</REQUESTDESC>
			</EXPORTDATA>
		</BODY>
	</ENVELOPE>`

	respBytes, err := a.postXML(ctx, xmlReq)
	if err != nil {
		return fmt.Errorf("Tally health check failed: %w", err)
	}

	if !strings.Contains(string(respBytes), "<ENVELOPE>") && !strings.Contains(string(respBytes), "<RESPONSE>") {
		return fmt.Errorf("invalid Tally health check response")
	}

	return nil
}

func (a *Adapter) ApplyPriceUpdate(ctx context.Context, barcode string, pricePaise int64) error {
	priceRupees := float64(pricePaise) / 100.0

	xmlReq := fmt.Sprintf(`<ENVELOPE>
		<HEADER>
			<TALLYREQUEST>Import Data</TALLYREQUEST>
		</HEADER>
		<BODY>
			<IMPORTDATA>
				<REQUESTDESC>
					<REPORTNAME>All Masters</REPORTNAME>
				</REQUESTDESC>
				<REQUESTDATA>
					<TALLYMESSAGE xmlns:UDF="TallyUDF">
						<STOCKITEM NAME="%s" ACTION="Alter">
							<MAILINGNAME>%s</MAILINGNAME>
							<RATE>%.2f/PCS</RATE>
							<OPENINGRATE>%.2f</OPENINGRATE>
						</STOCKITEM>
					</TALLYMESSAGE>
				</REQUESTDATA>
			</IMPORTDATA>
		</BODY>
	</ENVELOPE>`, barcode, barcode, priceRupees, priceRupees)

	respBytes, err := a.postXML(ctx, xmlReq)
	if err != nil {
		return fmt.Errorf("failed to apply price update to Tally: %w", err)
	}

	if strings.Contains(string(respBytes), "<LINEERROR>") {
		return fmt.Errorf("Tally error updating price: %s", string(respBytes))
	}

	a.logger.Info("[TallyAdapter] Applied price update barcode=%s priceRupees=%.2f", barcode, priceRupees)
	return nil
}

func (a *Adapter) ApplyStockAdjustment(ctx context.Context, barcode string, qtyDelta int64, reason string) error {
	xmlReq := fmt.Sprintf(`<ENVELOPE>
		<HEADER>
			<TALLYREQUEST>Import Data</TALLYREQUEST>
		</HEADER>
		<BODY>
			<IMPORTDATA>
				<REQUESTDESC>
					<REPORTNAME>Vouchers</REPORTNAME>
				</REQUESTDESC>
				<REQUESTDATA>
					<TALLYMESSAGE xmlns:UDF="TallyUDF">
						<VOUCHER VCHTYPE="Stock Journal" ACTION="Create">
							<DATE>%s</DATE>
							<NARRATION>Zippyra Stock Adjustment: %s</NARRATION>
							<INVENTORYENTRIES.LIST>
								<STOCKITEMNAME>%s</STOCKITEMNAME>
								<ISDEEMEDPOSITIVE>No</ISDEEMEDPOSITIVE>
								<ACTUALQTY>%d PCS</ACTUALQTY>
								<BILLEDQTY>%d PCS</BILLEDQTY>
							</INVENTORYENTRIES.LIST>
						</VOUCHER>
					</TALLYMESSAGE>
				</REQUESTDATA>
			</IMPORTDATA>
		</BODY>
	</ENVELOPE>`, time.Now().Format("20060102"), reason, barcode, qtyDelta, qtyDelta)

	respBytes, err := a.postXML(ctx, xmlReq)
	if err != nil {
		return fmt.Errorf("failed to apply stock adjustment to Tally: %w", err)
	}

	if strings.Contains(string(respBytes), "<LINEERROR>") {
		return fmt.Errorf("Tally error applying stock adjustment: %s", string(respBytes))
	}

	a.logger.Info("[TallyAdapter] Applied stock adjustment barcode=%s qtyDelta=%d reason=%s", barcode, qtyDelta, reason)
	return nil
}

func (a *Adapter) ApplyGrn(ctx context.Context, items []erp_adapter.GrnItem) error {
	var itemNodes strings.Builder
	for _, item := range items {
		costRupees := float64(item.CostPaise) / 100.0
		itemNodes.WriteString(fmt.Sprintf(`
			<INVENTORYENTRIES.LIST>
				<STOCKITEMNAME>%s</STOCKITEMNAME>
				<ISDEEMEDPOSITIVE>Yes</ISDEEMEDPOSITIVE>
				<RATE>%.2f/PCS</RATE>
				<ACTUALQTY>%d PCS</ACTUALQTY>
				<BILLEDQTY>%d PCS</BILLEDQTY>
			</INVENTORYENTRIES.LIST>`, item.Barcode, costRupees, item.Qty, item.Qty))
	}

	xmlReq := fmt.Sprintf(`<ENVELOPE>
		<HEADER>
			<TALLYREQUEST>Import Data</TALLYREQUEST>
		</HEADER>
		<BODY>
			<IMPORTDATA>
				<REQUESTDESC>
					<REPORTNAME>Vouchers</REPORTNAME>
				</REQUESTDESC>
				<REQUESTDATA>
					<TALLYMESSAGE xmlns:UDF="TallyUDF">
						<VOUCHER VCHTYPE="Purchase" ACTION="Create">
							<DATE>%s</DATE>
							<NARRATION>Zippyra GRN Inbound Sync</NARRATION>
							%s
						</VOUCHER>
					</TALLYMESSAGE>
				</REQUESTDATA>
			</IMPORTDATA>
		</BODY>
	</ENVELOPE>`, time.Now().Format("20060102"), itemNodes.String())

	respBytes, err := a.postXML(ctx, xmlReq)
	if err != nil {
		return fmt.Errorf("failed to apply GRN to Tally: %w", err)
	}

	if strings.Contains(string(respBytes), "<LINEERROR>") {
		return fmt.Errorf("Tally error applying GRN: %s", string(respBytes))
	}

	a.logger.Info("[TallyAdapter] Applied GRN with %d items", len(items))
	return nil
}

func (a *Adapter) PollLocalChanges(ctx context.Context, since time.Time) ([]erp_adapter.LocalChange, error) {
	xmlReq := `<ENVELOPE>
		<HEADER>
			<TALLYREQUEST>Export Data</TALLYREQUEST>
		</HEADER>
		<BODY>
			<EXPORTDATA>
				<REQUESTDESC>
					<REPORTNAME>List of Stock Items</REPORTNAME>
				</REQUESTDESC>
			</EXPORTDATA>
		</BODY>
	</ENVELOPE>`

	respBytes, err := a.postXML(ctx, xmlReq)
	if err != nil {
		return nil, fmt.Errorf("failed to poll local changes from Tally: %w", err)
	}

	type TallyItem struct {
		Name string `xml:"NAME,attr"`
		Rate string `xml:"RATE"`
	}
	type TallyExportResponse struct {
		XMLName xml.Name    `xml:"ENVELOPE"`
		Items   []TallyItem `xml:"BODY>DATA>TALLYMESSAGE>STOCKITEM"`
	}

	var parsed TallyExportResponse
	if err := xml.Unmarshal(respBytes, &parsed); err != nil {
		return []erp_adapter.LocalChange{}, nil
	}

	var changes []erp_adapter.LocalChange
	for _, item := range parsed.Items {
		if item.Name != "" {
			changes = append(changes, erp_adapter.LocalChange{
				EventType: "CATALOG_PRICE_CHANGED",
				Barcode:   item.Name,
				Payload: map[string]interface{}{
					"barcode": item.Name,
					"rate":    item.Rate,
				},
				Timestamp: time.Now(),
			})
		}
	}

	return changes, nil
}

func (a *Adapter) postXML(ctx context.Context, xmlPayload string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", a.endpoint, bytes.NewBufferString(xmlPayload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "text/xml;charset=utf-8")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Tally server returned status %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}
