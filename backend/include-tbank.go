// backend/include-tbank.go
package tbank

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Package tbank provides integration with T-Bank payment system
// Version: 1.0.0
// License: MIT

// Config represents T-Bank configuration
type Config struct {
	TerminalKey string `json:"terminalKey"` // Terminal key from T-Bank cabinet
	Password    string `json:"password"`    // Password for API
	BaseURL     string `json:"baseURL"`     // T-Bank API base URL
	CallbackURL string `json:"callbackURL"` // Callback URL for payment notifications
	SuccessURL  string `json:"successURL"`  // URL to redirect after success
	FailURL     string `json:"failURL"`     // URL to redirect after failure
}

// PaymentRequest represents payment initiation request
type PaymentRequest struct {
	OrderID     string  `json:"orderId"`     // Your internal order ID
	Amount      float64 `json:"amount"`      // Amount in rubles
	Description string  `json:"description"` // Payment description
	CustomerKey string  `json:"customerKey"` // Customer identifier (optional)
	Email       string  `json:"email"`       // Customer email (optional)
	Phone       string  `json:"phone"`       // Customer phone (optional)
	IP          string  `json:"ip"`          // Customer IP address
	PaymentType string  `json:"paymentType"` // Payment type: sbp, tpay, etc.
}

// PaymentResponse represents T-Bank API response
type PaymentResponse struct {
	Success     bool   `json:"Success"`
	ErrorCode   string `json:"ErrorCode"`
	TerminalKey string `json:"TerminalKey"`
	Status      string `json:"Status"`
	PaymentID   string `json:"PaymentId"`
	OrderID     string `json:"OrderId"`
	Amount      int    `json:"Amount"`
	PaymentURL  string `json:"PaymentURL"`
	Message     string `json:"Message"`
	Details     string `json:"Details"`
}

// PaymentStatus represents payment status
type PaymentStatus struct {
	OrderID   string    `json:"orderId"`
	PaymentID string    `json:"paymentId"`
	Status    string    `json:"status"`
	Amount    float64   `json:"amount"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// Notification represents T-Bank webhook notification
type Notification struct {
	TerminalKey string `json:"TerminalKey"`
	OrderID     string `json:"OrderId"`
	Success     bool   `json:"Success"`
	Status      string `json:"Status"`
	PaymentID   string `json:"PaymentId"`
	ErrorCode   string `json:"ErrorCode"`
	Amount      int    `json:"Amount"`
	CardID      int    `json:"CardId"`
	Pan         string `json:"Pan"`
	ExpDate     string `json:"ExpDate"`
	Token       string `json:"Token"`
}

// Client represents T-Bank API client
type Client struct {
	config     Config
	httpClient *http.Client
	cache      map[string]PaymentStatus
}

// NewClient creates new T-Bank client
func NewClient(config Config) *Client {
	return &Client{
		config: config,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		cache: make(map[string]PaymentStatus),
	}
}

// InitiatePayment initiates payment with T-Bank
func (c *Client) InitiatePayment(ctx context.Context, req PaymentRequest) (*PaymentResponse, error) {
	// Validate request
	if err := c.validatePaymentRequest(req); err != nil {
		return nil, err
	}

	// Convert amount to kopecks
	amountKop := int(req.Amount * 100)

	// Prepare request body for T-Bank
	requestBody := map[string]interface{}{
		"TerminalKey":     c.config.TerminalKey,
		"Amount":          amountKop,
		"OrderId":         req.OrderID,
		"Description":     req.Description,
		"SuccessURL":      c.config.SuccessURL,
		"FailURL":         c.config.FailURL,
		"NotificationURL": c.config.CallbackURL,
		"DATA": map[string]interface{}{
			"connection_type": "Widget",
			"PaymentType":     req.PaymentType,
		},
	}

	// Add optional fields
	if req.CustomerKey != "" {
		requestBody["CustomerKey"] = req.CustomerKey
	}
	if req.Email != "" {
		requestBody["Email"] = req.Email
	}
	if req.Phone != "" {
		requestBody["Phone"] = req.Phone
	}
	if req.IP != "" {
		requestBody["IP"] = req.IP
	}

	// Sign the request (simplified - in production use proper signing)
	requestBody["Token"] = c.generateToken(requestBody)

	// Make API request
	apiURL := c.config.BaseURL + "/Init"
	if c.config.BaseURL == "" {
		apiURL = "https://securepayments.tbank.ru/eacq/v2/Init"
	}

	response, err := c.makeRequest(ctx, "POST", apiURL, requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to initiate payment: %w", err)
	}

	// Cache payment status
	c.cache[req.OrderID] = PaymentStatus{
		OrderID:   req.OrderID,
		PaymentID: response.PaymentID,
		Status:    response.Status,
		Amount:    req.Amount,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	return response, nil
}

// GetPaymentStatus retrieves payment status from T-Bank
func (c *Client) GetPaymentStatus(ctx context.Context, orderID string) (*PaymentStatus, error) {
	// Check cache first
	if status, ok := c.cache[orderID]; ok {
		return &status, nil
	}

	// Prepare request to T-Bank
	requestBody := map[string]interface{}{
		"TerminalKey": c.config.TerminalKey,
		"OrderId":     orderID,
		"Token":       c.generateToken(map[string]interface{}{"TerminalKey": c.config.TerminalKey, "OrderId": orderID}),
	}

	apiURL := c.config.BaseURL + "/GetState"
	if c.config.BaseURL == "" {
		apiURL = "https://securepayments.tbank.ru/eacq/v2/GetState"
	}

	var response struct {
		Success   bool   `json:"Success"`
		ErrorCode string `json:"ErrorCode"`
		Status    string `json:"Status"`
		PaymentID string `json:"PaymentId"`
	}

	err := c.makeJSONRequest(ctx, "POST", apiURL, requestBody, &response)
	if err != nil {
		return nil, err
	}

	if !response.Success {
		return nil, fmt.Errorf("T-Bank error: %s", response.ErrorCode)
	}

	status := PaymentStatus{
		OrderID:   orderID,
		PaymentID: response.PaymentID,
		Status:    response.Status,
		UpdatedAt: time.Now(),
	}

	// Update cache
	c.cache[orderID] = status

	return &status, nil
}

// ConfirmPayment confirms payment (for two-step payments)
func (c *Client) ConfirmPayment(ctx context.Context, paymentID string, amount float64) error {
	amountKop := int(amount * 100)

	requestBody := map[string]interface{}{
		"TerminalKey": c.config.TerminalKey,
		"PaymentId":   paymentID,
		"Amount":      amountKop,
		"Token":       c.generateToken(map[string]interface{}{"TerminalKey": c.config.TerminalKey, "PaymentId": paymentID, "Amount": amountKop}),
	}

	apiURL := c.config.BaseURL + "/Confirm"
	if c.config.BaseURL == "" {
		apiURL = "https://securepayments.tbank.ru/eacq/v2/Confirm"
	}

	var response struct {
		Success   bool   `json:"Success"`
		ErrorCode string `json:"ErrorCode"`
	}

	err := c.makeJSONRequest(ctx, "POST", apiURL, requestBody, &response)
	if err != nil {
		return err
	}

	if !response.Success {
		return fmt.Errorf("confirm payment failed: %s", response.ErrorCode)
	}

	return nil
}

// CancelPayment cancels payment
func (c *Client) CancelPayment(ctx context.Context, paymentID string) error {
	requestBody := map[string]interface{}{
		"TerminalKey": c.config.TerminalKey,
		"PaymentId":   paymentID,
		"Token":       c.generateToken(map[string]interface{}{"TerminalKey": c.config.TerminalKey, "PaymentId": paymentID}),
	}

	apiURL := c.config.BaseURL + "/Cancel"
	if c.config.BaseURL == "" {
		apiURL = "https://securepayments.tbank.ru/eacq/v2/Cancel"
	}

	var response struct {
		Success   bool   `json:"Success"`
		ErrorCode string `json:"ErrorCode"`
	}

	err := c.makeJSONRequest(ctx, "POST", apiURL, requestBody, &response)
	if err != nil {
		return err
	}

	if !response.Success {
		return fmt.Errorf("cancel payment failed: %s", response.ErrorCode)
	}

	return nil
}

// HandleNotification processes T-Bank webhook notification
func (c *Client) HandleNotification(body []byte) (*Notification, error) {
	var notification Notification
	if err := json.Unmarshal(body, &notification); err != nil {
		return nil, err
	}

	// Verify notification (in production, verify signature/token)
	if notification.TerminalKey != c.config.TerminalKey {
		return nil, errors.New("invalid terminal key in notification")
	}

	// Update cache
	if status, ok := c.cache[notification.OrderID]; ok {
		status.Status = notification.Status
		status.UpdatedAt = time.Now()
		c.cache[notification.OrderID] = status
	}

	return &notification, nil
}

// validatePaymentRequest validates payment request
func (c *Client) validatePaymentRequest(req PaymentRequest) error {
	if req.OrderID == "" {
		return errors.New("orderId is required")
	}
	if req.Amount <= 0 {
		return errors.New("amount must be greater than 0")
	}
	if req.IP == "" {
		return errors.New("customer IP is required")
	}
	return nil
}

// generateToken generates request token (simplified version)
func (c *Client) generateToken(data map[string]interface{}) string {
	// In production, implement proper signing according to T-Bank documentation
	// This is a simplified example
	return "simplified_token_" + c.config.Password
}

// makeRequest makes HTTP request to T-Bank API
func (c *Client) makeRequest(ctx context.Context, method, url string, body interface{}) (*PaymentResponse, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var paymentResp PaymentResponse
	if err := json.Unmarshal(respBody, &paymentResp); err != nil {
		return nil, err
	}

	return &paymentResp, nil
}

// makeJSONRequest makes request and unmarshals response
func (c *Client) makeJSONRequest(ctx context.Context, method, url string, body, result interface{}) error {
	resp, err := c.makeRequest(ctx, method, url, body)
	if err != nil {
		return err
	}

	// Convert PaymentResponse to target struct
	data, err := json.Marshal(resp)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, result)
}

// HTTP Handler for frontend integration
func (c *Client) PaymentHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			var req struct {
				PaymentType string `json:"paymentType"`
				TerminalKey string `json:"terminalKey"`
			}

			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			// In production, get order details from your database
			orderID := fmt.Sprintf("order_%d", time.Now().Unix())

			paymentReq := PaymentRequest{
				OrderID:     orderID,
				Amount:      1.50, // Example amount
				Description: "Payment for order",
				IP:          r.RemoteAddr,
				PaymentType: req.PaymentType,
			}

			resp, err := c.InitiatePayment(r.Context(), paymentReq)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success":    resp.Success,
				"paymentUrl": resp.PaymentURL,
				"orderId":    orderID,
				"paymentId":  resp.PaymentID,
			})

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

// GetConfig returns current configuration (without sensitive data)
func (c *Client) GetConfig() map[string]interface{} {
	return map[string]interface{}{
		"terminalKey": c.config.TerminalKey,
		"baseURL":     c.config.BaseURL,
		"callbackURL": c.config.CallbackURL,
		"successURL":  c.config.SuccessURL,
		"failURL":     c.config.FailURL,
	}
}

// ClearCache clears internal cache
func (c *Client) ClearCache() {
	c.cache = make(map[string]PaymentStatus)
}
