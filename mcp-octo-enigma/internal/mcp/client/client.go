package client

import "context"

type Client struct{}

func New() *Client { return &Client{} }
func (c *Client) Connect(ctx context.Context) error { _ = ctx; return nil }
