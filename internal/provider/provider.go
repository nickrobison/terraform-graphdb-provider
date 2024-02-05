// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure GraphDBProvider satisfies various provider interfaces.
var _ provider.Provider = &GraphDBProvider{}

// GraphDBProvider defines the provider implementation.
type GraphDBProvider struct {
	version string
}

// GraphDBProviderModel describes the provider data model.
type GraphDBProviderModel struct {
	Host     types.String `tfsdk:"host"`
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
	Port     types.Int64  `tfsdk:"port"`
}

const (
	providerName = "graphdb"
)

func (p *GraphDBProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = providerName
	resp.Version = p.version
}

func (p *GraphDBProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Optional: true,
			},
			"username": schema.StringAttribute{
				Optional: true,
				Description: "This is the username for the API connection." +
					" May also be provided via " + EnvUsername + " environment variable.",
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"password": schema.StringAttribute{
				Optional:  true,
				Sensitive: true,
				Description: "This is the password for the API connection." +
					" May also be provided via " + EnvPassword + " environment variable.",
			},
			"port": schema.Int64Attribute{
				Optional: true,
				Description: "This is the tcp port for the API connection." +
					" May also be provided via " + EnvPort + " environment variable.",
				Validators: []validator.Int64{
					int64validator.Between(1, 65535),
				},
			},
		},
	}
}

func (p *GraphDBProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Initializing GraphDB provider")
	var config GraphDBProviderModel

	diags := req.Config.Get(ctx, &config)

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	unknownValueErrorMessage := "The provider cannot create the GraphDB client as there is an unknown configuration value "
	instructionUnknownMessage := " Either target apply the source of the value first, " +
		"set the value statically in the configuration, or use the %s environment variable."

	if config.Host.IsUnknown() {
		resp.Diagnostics.AddAttributeError(path.Root("address"), "Unknown GraphDB Host", fmt.Sprintf("%s for the GraphDB Host. %s", unknownValueErrorMessage, fmt.Sprintf(instructionUnknownMessage, EnvHost)))
	}

	if config.Port.IsUnknown() {
		resp.Diagnostics.AddAttributeError(path.Root("port"), "Unknown GraphDB Port", fmt.Sprintf("%s for the GraphDB Port. %s", unknownValueErrorMessage, fmt.Sprintf(instructionUnknownMessage, EnvPort)))
	}

	if config.Username.IsUnknown() {
		resp.Diagnostics.AddAttributeError(path.Root("username"), "Unknown GraphDB Username", fmt.Sprintf("%s for the GraphDB Username. %s", unknownValueErrorMessage, fmt.Sprintf(instructionUnknownMessage, EnvUsername)))
	}

	if config.Password.IsUnknown() {
		resp.Diagnostics.AddAttributeError(path.Root("password"), "Unknown GraphDB Password", fmt.Sprintf("%s for the GraphDB Password. %s", unknownValueErrorMessage, fmt.Sprintf(instructionUnknownMessage, EnvPassword)))
	}

	if resp.Diagnostics.HasError() {
		return
	}

	host := os.Getenv(EnvHost)
	if !config.Host.IsNull() {
		host = config.Host.ValueString()
	}
	if host == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("address"),
			"Missing GraphDB address target",
			"The provider cannot create the GraphDB client as there is a missing or empty value for the GraphDB address."+
				" Set the value in the configuration or use the "+EnvHost+" environment variable."+
				" If either is already set, ensure the value is not empty.",
		)

		return
	}

	client := NewClient(host)

	if !config.Port.IsNull() {
		client.WithPort(int(config.Port.ValueInt64()))
	} else if v := os.Getenv(EnvPort); v != "" {
		d, err := strconv.Atoi(v)
		if err != nil {
			resp.Diagnostics.AddAttributeWarning(
				path.Root("port"),
				"Failed to parse "+EnvPort,
				fmt.Sprintf("Error to parse value in "+EnvPort+" environment variable: %s\n"+
					"So the variable is not used", err),
			)
		} else {
			client.WithPort(d)
		}
	}

	var username = ""

	if !config.Username.IsNull() {
		username = config.Username.ValueString()

	} else if v := os.Getenv(EnvUsername); v != "" {
		username = v
	}
	client.WithUsername(username)

	var password = ""
	if !config.Password.IsNull() {
		password = config.Password.ValueString()
	} else if v := os.Getenv(EnvPassword); v != "" {
		password = v
	}
	client.WithPassword(password)

	ctx = tflog.SetField(ctx, "graphdb_host", host)
	ctx = tflog.SetField(ctx, "graphdb_username", username)
	ctx = tflog.MaskFieldValuesWithFieldKeys(ctx, "graphdb_password", password)
	tflog.Info(ctx, "Created client")

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *GraphDBProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewRepositoryResource,
		NewUserResource,
	}
}

func (p *GraphDBProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewRepositoriesDataSource,
		NewUserDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &GraphDBProvider{
			version: version,
		}
	}
}

func unexpectedDataSourceConfigureType(
	_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse,
) {
	resp.Diagnostics.AddError(
		"Unexpected Data Source Configure Type",
		fmt.Sprintf(
			"Expected *Client, got: %T. Please report this issue to the provider developers.",
			req.ProviderData,
		),
	)
}
