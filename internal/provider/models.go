// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

type repositoryListResponse struct {
	Name        string `json:"name"`
	Title       string `json:"title"`
	Uri         string `json:"uri"`
	ExternalUrl string `json:"external_url"`
	Type        string `json:"type"`
	Local       bool   `json:"local"`
}

type repositoryGetResponse struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Type     string `json:"type"`
	Location string `json:"location"`
}

type userCreateRequest struct {
	Username    string   `json:"username"`
	Password    string   `json:"password"`
	Authorities []string `json:"grantedAuthorities"`
}

type userGetResponse struct {
	Username    string   `json:"username"`
	Password    string   `json:"password"`
	Authorities []string `json:"grantedAuthorities"`
}
