package trackdrive

import "net/http"

type Client struct {
	authHeader string
	httpClient *http.Client
}

func New(authHeader string, httpClient *http.Client) *Client {
	return &Client{
		authHeader: authHeader,
		httpClient: httpClient,
	}
}
