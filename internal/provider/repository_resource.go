// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource = &RepositoryResource{}
)

type RepositoryResource struct {
	client *Client
}

type RepositoryResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Config      types.String `tfsdk:"config"`
	Description types.String `tfsdk:"description"`
	Location    types.String `tfsdk:"location"`
	Type        types.String `tfsdk:"type"`
}

func NewRepositoryResource() resource.Resource {
	return &RepositoryResource{}
}

func (r *RepositoryResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*Client)
	if !ok {
		ProviderDataError(req.ProviderData, &resp.Diagnostics)
		return
	}

	r.client = client
}

func (r *RepositoryResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_repository"
}

func (r *RepositoryResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "GraphDB Repository",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "An identifier for the resource",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Repository name",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "Repository Description",
			},
			"config": schema.StringAttribute{
				Optional:    true,
				Description: "Configuration file in Turtle syntax",
			},
			"location": schema.StringAttribute{
				Computed:    true,
				Description: "Repository External URL"},
			"type": schema.StringAttribute{
				Computed:    true,
				Description: "Type of Repository"},
		},
	}
}

func (r *RepositoryResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan RepositoryResourceModel

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.Config.ValueString() == "" {
		resp.Diagnostics.AddAttributeError(path.Root("config"), "Empty Config", "Config cannot be empty on creation.")
		return
	}
	reader := strings.NewReader(plan.Config.ValueString())

	var err = r.client.CreateRepository(ctx, reader)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create Repository", fmt.Sprintf("Failed to create repository. Unexpected error %s", err.Error()))
		return
	}
	// Read the repository back out
	// TODO: This is unsafe because there could be a mismatch between the config file and the repo name.
	repo, err := r.client.GetRepository(ctx, plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to create Repository", fmt.Sprintf("Failed to retrieve repository after creation. Unexpected error %s", err.Error()))
		return
	}

	plan.ID = types.StringValue(repo.ID)
	plan.Description = types.StringValue(repo.Title)
	plan.Type = types.StringValue(repo.Type)
	plan.Location = types.StringValue(repo.Location)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

func (r *RepositoryResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {

	var state RepositoryResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	repoName := state.Name.ValueString()
	tflog.Debug(ctx, "Fetching repository", map[string]any{"id": repoName})
	err := r.doRead(ctx, repoName, &state)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read repository", fmt.Sprintf("Unable to read repository. Unexpected error: %a", err))
		return
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *RepositoryResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state RepositoryResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	repoId := state.ID.ValueString()
	tflog.Debug(ctx, "Deleting repository", map[string]any{"id": repoId})

	err := r.client.DeleteRepository(ctx, repoId)
	if err != nil {
		resp.Diagnostics.AddError("Error deleting repository", fmt.Sprintf("Could not delete repository. Unexpected error: %s", err.Error()))
		return
	}
}

func (r *RepositoryResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// TODO: Implement
}

func (r *RepositoryResource) doRead(ctx context.Context, id string, data *RepositoryResourceModel) error {
	repo, err := r.client.GetRepository(ctx, id)
	if err != nil {
		return err
	}

	data.ID = types.StringValue(repo.ID)
	data.Name = types.StringValue(repo.ID)
	data.Description = types.StringValue(repo.Title)
	data.Type = types.StringValue(repo.Type)
	return nil
}
