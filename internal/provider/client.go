// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
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

func (c *Client) GetRepositories(ctx context.Context) ([]repositoryListResponse, error) {
	var data = []repositoryListResponse{}

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

func (c *Client) CreateRepository(ctx context.Context, input string) error {

	body := &bytes.Buffer{}

	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormField("config")
	if err != nil {
		return err
	}
	_, err = io.Copy(part, strings.NewReader(input))
	if err != nil {
		return err
	}
	writer.Close()

	req, err := http.NewRequestWithContext(ctx, "POST", c.createUrl("repositories"), body)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", writer.FormDataContentType())
	_, err = c.doRequest(req)
	// TODO: Check return code
	return err
}

func (c *Client) GetRepository(ctx context.Context, id string) (repositoryGetResponse, error) {
	var data = repositoryGetResponse{}

	req, err := http.NewRequestWithContext(ctx, "GET", c.createUrl("/repository/"+id), nil)
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

func (c *Client) DeleteRepository(ctx context.Context, id string) error {
	req, err := http.NewRequestWithContext(ctx, "DELETE", c.createUrl("/repository/"+id), nil)
	if err != nil {
		return err
	}

	_, err = c.doRequest(req)
	// TODO: Check response code
	return err
}

func (c *Client) doRequest(req *http.Request) (*http.Response, error) {
	req.SetBasicAuth(c.username, c.password)
	return c.client.Do(req)
}

func (c *Client) createUrl(resource string) string {
	return fmt.Sprintf("%s:%d/rest/%s", c.address, c.port, resource)
}
