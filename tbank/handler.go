package tbank

import (
	"encoding/json"
	"net/http"
)

type InitHTTPRequest struct {
	OrderID     string `json:"orderId"`
	Amount      int64  `json:"amount"` // копейки
	Description string `json:"description"`
}

func InitHandler(c *Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req InitHTTPRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}

		resp, err := c.Init(r.Context(), InitRequest(req))
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}
