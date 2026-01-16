package tbank

const baseURL = "https://securepayments.tbank.ru/eacq/v2"

type Client struct {
	cfg Config
}

func New(cfg Config) *Client {
	return &Client{cfg: cfg}
}
