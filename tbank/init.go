package tbank

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type InitRequest struct {
	OrderID     string
	Amount      int64 // в копейках
	Description string
}

type InitResponse struct {
	PaymentID  string `json:"PaymentId"`
	PaymentURL string `json:"PaymentURL"`
}

func (c *Client) Init(ctx context.Context, r InitRequest) (*InitResponse, error) {
	body := map[string]string{
		"TerminalKey":     c.cfg.TerminalKey,
		"OrderId":         r.OrderID,
		"Amount":          fmt.Sprintf("%d", r.Amount),
		"Description":     r.Description,
		"SuccessURL":      c.cfg.SuccessURL,
		"FailURL":         c.cfg.FailURL,
		"NotificationURL": c.cfg.CallbackURL,
	}

	body["Token"] = generateToken(body, c.cfg.Password)

	b, _ := json.Marshal(body)

	req, _ := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		baseURL+"/Init",
		bytes.NewReader(b),
	)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var raw struct {
		Success bool   `json:"Success"`
		Message string `json:"Message"`
		InitResponse
	}

	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}

	if !raw.Success {
		return nil, fmt.Errorf("tbank error: %s", raw.Message)
	}

	return &raw.InitResponse, nil
}
