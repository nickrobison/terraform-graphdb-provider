// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &repositoryDataSource{}
	_ datasource.DataSourceWithConfigure = &repositoryDataSource{}
)

type repositoryDataSource struct {
	client *Client
}

type repositoriesDataSourceModel struct {
	Repositories []repositoryDataModel `tfsdk:"repositories"`
}

type repositoryDataModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Uri         types.String `tfsdk:"uri"`
	ExternalUrl types.String `tfsdk:"external_url"`
	Type        types.String `tfsdk:"type"`
	Local       types.Bool   `tfsdk:"local"`
}

func NewRepositoriesDataSource() datasource.DataSource {
	return &repositoryDataSource{}
}

func (d *repositoryDataSource) Metadata(
	_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_repositories"
}

func (d *repositoryDataSource) Configure(
	ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*Client)
	if !ok {
		unexpectedDataSourceConfigureType(ctx, req, resp)
	}
	d.client = client
}

func (d *repositoryDataSource) Schema(
	_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Description: "Get configuration from a Repository.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "An identifier for the data source",
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "Name of the Repository"},
			"description": schema.StringAttribute{
				Computed:    true,
				Optional:    true,
				Description: "Repository Description",
			},
			"uri": schema.StringAttribute{
				Computed:    true,
				Description: "Repository URI"},
			"externalUrl": schema.StringAttribute{
				Computed:    true,
				Description: "Repository External URL"},
			"type": schema.StringAttribute{
				Computed:    true,
				Description: "Type of Repository"},
			"local": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether repository is local to the GraphDB Instance"},
		},
	}
}

func (d *repositoryDataSource) Read(
	ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse,
) {
	var state repositoriesDataSourceModel

	repositories, err := d.client.GetRepositories(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to get Repositories", err.Error())
		return
	}

	for _, repo := range repositories {
		repoState := repositoryDataModel{
			ID:          types.StringValue(repo.Name),
			Name:        types.StringValue(repo.Name),
			Description: types.StringValue(repo.Title),
			Uri:         types.StringValue(repo.Uri),
			ExternalUrl: types.StringValue(repo.ExternalUrl),
			Type:        types.StringValue(repo.Type),
			Local:       types.BoolValue(repo.Local),
		}

		state.Repositories = append(state.Repositories, repoState)
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}
