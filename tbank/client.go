package tbank

const baseURL = "https://securepay.tinkoff.ru/v2"

type Client struct {
	cfg Config
}

func New(cfg Config) *Client {
	return &Client{cfg: cfg}
}
