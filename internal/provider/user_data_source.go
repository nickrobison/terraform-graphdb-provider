package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &userDataSource{}
	_ datasource.DataSourceWithConfigure = &userDataSource{}
)

type userDataSource struct {
	client *Client
}

type usersDataSourceModel struct {
	Users []userDataModel `tfsdk:"users"`
}

type userDataModel struct {
	ID       types.String `tfsdk:"id"`
	Username types.String `tfsdk:"username"`
	Role     types.String `tfsdk:"role"`
}

func NewUserDataSource() datasource.DataSource {
	return &userDataSource{}
}

func (d *userDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_users"
}

func (d *userDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*Client)
	if !ok {
		unexpectedDataSourceConfigureType(ctx, req, resp)
	}
	d.client = client
}

func (d *userDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Get users from a GraphDB database",
		Attributes: map[string]schema.Attribute{
			"users": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{

						"id": schema.StringAttribute{
							Computed:    true,
							Description: "An identifier for the resource",
						},
						"username": schema.StringAttribute{
							Optional:    true,
							Description: "Username",
						},
						"role": schema.StringAttribute{
							Optional:    true,
							Description: "User role",
						},
					},
				},
			},
		},
	}
}

func (d *userDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state usersDataSourceModel

	users, err := d.client.GetUsers(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to get users", err.Error())
		return
	}

	for _, user := range users {
		r, err := authorityToRole(user.Authorities[0])
		if err != nil {
			resp.Diagnostics.AddError("Failed to retrieve user", fmt.Sprintf("Unknown role %s for user %s", user.Authorities[0], user.Username))
			continue
		}
		userState := userDataModel{
			ID:       types.StringValue(user.Username),
			Username: types.StringValue(user.Username),
			Role:     types.StringValue(r),
		}
		state.Users = append(state.Users, userState)
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
