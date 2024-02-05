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
	"os"
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

	part, err := writer.CreateFormFile("config", input)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(input, os.O_RDONLY, 0600)
	if err != nil {
		return err
	}
	_, err = io.Copy(part, f)
	if err != nil {
		return err
	}

	writer.Close()

	req, err := http.NewRequestWithContext(ctx, "POST", c.createUrl("repositories"), body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.doRequest(req)
	if resp.StatusCode != 201 {
		defer resp.Body.Close()
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Failed to create repository: %s", string(b))
	}
	return err
}

func (c *Client) GetRepository(ctx context.Context, id string) (repositoryGetResponse, error) {
	var data = repositoryGetResponse{}

	req, err := http.NewRequestWithContext(ctx, "GET", c.createUrl("repositories/"+id), nil)
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
	req, err := http.NewRequestWithContext(ctx, "DELETE", c.createUrl("repositories/"+id), nil)
	if err != nil {
		return err
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Failed to delete repository %s. error: %s", id, string(b))
	}

	return nil
}

func (c *Client) CreateUser(ctx context.Context, create userCreateRequest) error {
	body, err := json.Marshal(create)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, "POST", c.createUrl("security/users/"+create.Username), bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	resp, err := c.doRequest(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusCreated {
		defer resp.Body.Close()
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Failed to create user. Error: %s", string(b))
	}
	return nil
}

func (c *Client) GetUsers(ctx context.Context) ([]userGetResponse, error) {
	var users []userGetResponse

	req, err := http.NewRequestWithContext(ctx, "GET", c.createUrl("security/users/"), nil)
	if err != nil {
		return users, err
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return users, err
	}

	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&users)
	return users, err

}

func (c *Client) GetUser(ctx context.Context, username string) (userGetResponse, error) {
	var user userGetResponse

	req, err := http.NewRequestWithContext(ctx, "GET", c.createUrl("security/users/"+username), nil)
	if err != nil {
		return user, err
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return user, err
	}
	if resp.StatusCode == http.StatusNotFound {
		return user, fmt.Errorf("Cannot find user with name: %s", username)
	}

	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&user)
	return user, err

}

func (c *Client) UpdateUser(ctx context.Context, username string, update userCreateRequest) error {
	body, err := json.Marshal(update)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, "PUT", c.createUrl("security/users/"+username), bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	resp, err := c.doRequest(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Failed to update user. Error: %s", string(b))
	}
	return nil
}

func (c *Client) DeleteUser(ctx context.Context, username string) error {
	req, err := http.NewRequestWithContext(ctx, "DELETE", c.createUrl("security/users/"+username), nil)
	if err != nil {
		return err
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusNoContent {
		defer resp.Body.Close()
		b, _ := io.ReadAll(resp.Request.Body)
		return fmt.Errorf("Failed to delete user: %s. Error: %s", username, string(b))
	}
	return nil
}

func (c *Client) doRequest(req *http.Request) (*http.Response, error) {
	req.SetBasicAuth(c.username, c.password)
	req.Header.Set("Content-Type", "application/json")
	return c.client.Do(req)
}

func (c *Client) createUrl(resource string) string {
	return fmt.Sprintf("http://%s:%d/rest/%s", c.address, c.port, resource)
}
