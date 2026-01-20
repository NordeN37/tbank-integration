package tbank

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/NordeN37/tbank-integration/http2curl"
)

type InitRequestReceipt struct {
	Email    string                    `json:"Email"`
	Phone    string                    `json:"Phone"`
	Taxation string                    `json:"Taxation"`
	Items    []InitRequestReceiptItems `json:"Items"`
}
type InitRequestReceiptItems struct {
	Name     string  `json:"Name"`
	Price    int     `json:"Price"`
	Quantity int     `json:"Quantity"`
	Amount   int     `json:"Amount"`
	Tax      string  `json:"Tax"`
	Ean13    *string `json:"Ean13,omitempty"`
}
type InitRequestReceiptsDATA struct {
	ConnectionType string `json:"connection_type"`
}
type InitRequest struct {
	Amount      int                      `json:"Amount"`
	OrderId     string                   `json:"OrderId"`
	Description string                   `json:"Description"`
	Receipt     InitRequestReceipt       `json:"Receipt"`
	DATA        *InitRequestReceiptsDATA `json:"DATA"`
}
type InitResponse struct {
	Success     bool   `json:"Success"`
	ErrorCode   string `json:"ErrorCode"`
	TerminalKey string `json:"TerminalKey"`
	Status      string `json:"Status"`
	PaymentId   string `json:"PaymentId"`
	OrderId     string `json:"OrderId"`
	Amount      int    `json:"Amount"`
	PaymentURL  string `json:"PaymentURL"`
	Message     string `json:"Message"`
	Details     string `json:"Details"`
}

func (c *Client) Init(ctx context.Context, r InitRequest) (*InitResponse, error) {
	body := map[string]interface{}{
		"TerminalKey": c.cfg.TerminalKey,
		"OrderId":     r.OrderId,
		"Amount":      fmt.Sprintf("%d", r.Amount),
		"Description": r.Description,
	}
	body["Token"] = generateToken(body, c.cfg.SecretKey)

	// Добавляем Receipt если он есть
	if r.Receipt.Email != "" || r.Receipt.Phone != "" {
		// Преобразуем Receipt в нужный формат
		receiptData := map[string]interface{}{
			"Email":    r.Receipt.Email,
			"Phone":    r.Receipt.Phone,
			"Taxation": r.Receipt.Taxation,
		}

		// Преобразуем Items
		items := make([]map[string]interface{}, len(r.Receipt.Items))
		for i, item := range r.Receipt.Items {
			items[i] = map[string]interface{}{
				"Name":     item.Name,
				"Price":    item.Price,
				"Quantity": item.Quantity,
				"Amount":   item.Amount,
				"Tax":      item.Tax,
			}
		}
		receiptData["Items"] = items
		body["Receipt"] = receiptData
	}

	b, _ := json.Marshal(body)

	req, _ := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		baseURL+"/Init",
		bytes.NewReader(b),
	)
	req.Header.Set("Content-Type", "application/json")
	if c.cfg.Debug {
		curlCommand, err := http2curl.GetCurlCommand(req)
		if err != nil {
			return nil, err
		}
		log.Println(curlCommand.String())
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var raw InitResponse

	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}

	if !raw.Success {
		return nil, fmt.Errorf("tbank error: %s; details %s", raw.Message, raw.Details)
	}

	return &raw, nil
}
