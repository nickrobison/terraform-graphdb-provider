// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

type Client struct {
	address  string
	username string
	password string
	port     int
}

func NewClient(address string) *Client {
	return &Client{
		address:  address,
		port:     7200,
		username: "",
		password: "",
	}
}

func (c *Client) WithPort(port int) *Client {
	c.port = port
	return c
}

func (c *Client) WithUsername(username string) *Client {
	c.username = username
	return c
}

func (c *Client) WithPassword(password string) *Client {
	c.password = password
	return c
}
