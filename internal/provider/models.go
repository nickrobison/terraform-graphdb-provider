// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

type repositoryResponse struct {
	Name        string `json:"name"`
	Title       string `json:"title"`
	Uri         string `json:"uri"`
	ExternalUrl string `json:"external_url"`
	Type        string `json:"type"`
	Local       bool   `json:"local"`
}
