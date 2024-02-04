// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &UserResource{}
	_ resource.ResourceWithImportState = &UserResource{}
)

type UserResource struct {
	client *Client
}

type UserResourceModel struct {
	ID       types.String `tfsdk:"id"`
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
	Role     types.String `tfsdk:"role"`
}

func NewUserResource() resource.Resource {
	return &UserResource{}
}

func (r *UserResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*Client)
	if !ok {
		ProviderDataError(req.ProviderData, &resp.Diagnostics)
	}

	r.client = client
}

func (r *UserResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (r *UserResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "An identifier for the resource",
			},
			"username": schema.StringAttribute{
				Required:    true,
				Description: "Username",
			},
			"password": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Optional user password",
			},
			"role": schema.StringAttribute{
				Required:    true,
				Description: "User role",
				Validators: []validator.String{
					stringvalidator.OneOf("user", "repo-manager", "admin"),
				},
			},
		},
	}
}

func (r *UserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan UserResourceModel

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	username := plan.Username.ValueString()

	request := userCreateRequest{
		Username: username,
		Password: plan.Password.ValueString(),
		Authorities: []string{
			roleToAuthority(plan.Role.ValueString()),
		},
	}
	tflog.Debug(ctx, "Attempting to create user", map[string]any{"username": username})

	err := r.client.CreateUser(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create User", fmt.Sprintf("Failed to create user. Unexpected error: %s", err.Error()))
		return
	}

	// Read the user back out, since the GraphDB API doesn't return the payload back to us
	err = r.doRead(ctx, plan.Username.ValueString(), &plan)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create User", fmt.Sprintf("Failed to retrieve user after creation. Unexpected error: %s", err.Error()))
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

func (r *UserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state UserResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	username := state.ID.ValueString()
	tflog.Debug(ctx, "Reading user", map[string]any{"username": username})
	err := r.doRead(ctx, username, &state)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read user", fmt.Sprintf("Unable to read user. Unexpected error: %s", err.Error()))
	}
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *UserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan UserResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	username := plan.Username.ValueString()
	request := userCreateRequest{
		Username:    username,
		Password:    plan.Password.ValueString(),
		Authorities: []string{roleToAuthority(plan.Role.ValueString())},
	}
	tflog.Debug(ctx, "Updating user", map[string]any{"username": username})

	err := r.client.UpdateUser(ctx, username, request)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update User", fmt.Sprintf("Failed to update user. Unexpected error: %s", err.Error()))
		return
	}
	err = r.doRead(ctx, username, &plan)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update User", fmt.Sprintf("Failed to retrieve user after update. Unexpected error: %s", err.Error()))
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

func (r *UserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state UserResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	username := state.ID.ValueString()
	tflog.Debug(ctx, "Deleting user", map[string]any{"username": username})
	err := r.client.DeleteUser(ctx, username)
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete user", fmt.Sprintf("Could not delete user. Unexpected error: %s", err.Error()))
	}

}

func (r *UserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *UserResource) doRead(ctx context.Context, username string, data *UserResourceModel) error {
	user, err := r.client.GetUser(ctx, username)
	if err != nil {
		return err
	}

	role, err := authorityToRole(user.Authorities[0])
	if err != nil {
		return err
	}

	data.ID = types.StringValue(user.Username)
	data.Username = types.StringValue(user.Username)
	data.Role = types.StringValue(role)
	return nil
}
