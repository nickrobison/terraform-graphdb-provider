// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type Client struct {
	client   *http.Client
	address  string
	username string
	password string
	port     int
}

func NewClient(address string) *Client {
	return &Client{
		client:   &http.Client{},
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

func (c *Client) GetRepositories(ctx context.Context) ([]repositoryResponse, error) {
	var data = []repositoryResponse{}

	req, err := http.NewRequestWithContext(ctx, "GET", c.createUrl("repositories"), nil)
	if err != nil {
		return data, err
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return data, err
	}
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&data)
	return data, err
}

func (c *Client) doRequest(req *http.Request) (*http.Response, error) {
	req.SetBasicAuth(c.username, c.password)
	return c.client.Do(req)
}

func (c *Client) createUrl(resource string) string {
	return fmt.Sprintf("%s:%d/rest/%s", c.address, c.port, resource)
}
