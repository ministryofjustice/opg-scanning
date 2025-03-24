package api

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ministryofjustice/opg-scanning/internal/httpclient"
)

type Client struct {
	Middleware *httpclient.Middleware
}

func NewClient(middleware *httpclient.Middleware) *Client {
	return &Client{
		Middleware: middleware,
	}
}

func (c *Client) ClientRequest(ctx context.Context, reqData interface{}, url string) (*[]byte, error) {
	if reqData == nil {
		return nil, fmt.Errorf("request data is nil")
	}

	body, err := json.Marshal(reqData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request data: %w", err)
	}

	responseBody, err := c.Middleware.HTTPRequest(ctx, url, "POST", body, nil)
	if err != nil {
		return nil, fmt.Errorf("request to Sirius API (%v) failed: %w", url, err)
	}

	return &responseBody, nil
}
