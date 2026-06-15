// Copyright (c) Vates
// SPDX-License-Identifier: Apache-2.0

package datasources

import (
	"context"
	"fmt"

	"github.com/gofrs/uuid"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vatesfr/terraform-provider-xenorchestra/v2/internal/helpers"
	"github.com/vatesfr/xenorchestra-go-sdk/pkg/payloads"
	v2 "github.com/vatesfr/xenorchestra-go-sdk/v2"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &poolDataSource{}

func NewPoolDataSource() datasource.DataSource {
	return &poolDataSource{}
}

type poolDataSource struct {
	client *v2.XOClient
}

type poolDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	NameLabel   types.String `tfsdk:"name_label"`
	Description types.String `tfsdk:"description"`
	Master      types.String `tfsdk:"master"`
	DefaultSR   types.String `tfsdk:"default_sr"`
	Cores       types.Int64  `tfsdk:"cores"`
	Sockets     types.Int64  `tfsdk:"sockets"`
	Tags        types.List   `tfsdk:"tags"`
}

func poolModelFromPayload(ctx context.Context, pool *payloads.Pool) (poolDataSourceModel, diag.Diagnostics) {
	tags, diags := types.ListValueFrom(ctx, types.StringType, pool.Tags)

	result := poolDataSourceModel{
		ID:          types.StringValue(pool.ID.String()),
		NameLabel:   types.StringValue(pool.NameLabel),
		Description: types.StringValue(pool.NameDescription),
		Master:      types.StringValue(pool.Master.String()),
		DefaultSR:   types.StringValue(pool.DefaultSR.String()),
		Cores:       types.Int64Value(int64(pool.CPUs.Cores)),
		Sockets:     types.Int64Value(int64(pool.CPUs.Sockets)),
		Tags:        tags,
	}

	return result, diags
}

func poolSchemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			MarkdownDescription: "The id of the pool.",
			Computed:            true,
		},
		"name_label": schema.StringAttribute{
			MarkdownDescription: "The name label of the pool.",
			Computed:            true,
		},
		"description": schema.StringAttribute{
			MarkdownDescription: "The description of the pool.",
			Computed:            true,
		},
		"master": schema.StringAttribute{
			MarkdownDescription: "The id of the primary instance in the pool.",
			Computed:            true,
		},
		"default_sr": schema.StringAttribute{
			MarkdownDescription: "The default storage repository for the pool.",
			Computed:            true,
		},
		"cores": schema.Int64Attribute{
			MarkdownDescription: "Number of CPU cores in the pool.",
			Computed:            true,
		},
		"sockets": schema.Int64Attribute{
			MarkdownDescription: "Number of CPU sockets in the pool.",
			Computed:            true,
		},
		"tags": schema.ListAttribute{
			MarkdownDescription: "The tags associated with the pool.",
			Computed:            true,
			ElementType:         types.StringType,
		},
	}
}

func (d *poolDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_pool"
}

func (d *poolDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	// Set "id" to required instead of computed for the single pool data source
	attrs := poolSchemaAttributes()
	attrs["id"] = schema.StringAttribute{
		MarkdownDescription: attrs["id"].GetMarkdownDescription() + " Exactly one of `id` or `name_label` must be provided.",
		// Required:            true,
		Optional: true,
		Computed: false,
		Validators: []validator.String{
			stringvalidator.ExactlyOneOf(path.MatchRoot("id"), path.MatchRoot("name_label")),
			stringvalidator.LengthAtLeast(1),
		},
	}
	attrs["name_label"] = schema.StringAttribute{
		MarkdownDescription: attrs["name_label"].GetMarkdownDescription() + " Exactly one of `id` or `name_label` must be provided.",
		Computed:            true,
		Optional:            true,
		// Required:            true,
		Validators: []validator.String{
			stringvalidator.ExactlyOneOf(path.MatchRoot("id"), path.MatchRoot("name_label")),
			stringvalidator.LengthAtLeast(1),
		},
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: "Use this data source to get information about a Xen Orchestra pool.",
		Attributes:          attrs,
	}
}

func (d *poolDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = helpers.ProviderDataToXOClient(req.ProviderData, &resp.Diagnostics)
}

func (d *poolDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data poolDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var pool *payloads.Pool

	if data.ID.IsUnknown() || data.ID.IsNull() {
		// Retrieve pool by name_label if id is not provided
		poolName := data.NameLabel.ValueString()
		pools, err := d.client.Pool().GetAll(ctx, 0, fmt.Sprintf("name_label:\"%s\"", poolName))
		if err != nil {
			resp.Diagnostics.AddError(
				"Error reading pools",
				fmt.Sprintf("Could not read pools to find pool with name_label %s: %s", poolName, err.Error()),
			)
			return
		}
		if len(pools) == 0 {
			resp.Diagnostics.AddError(
				"Pool not found",
				fmt.Sprintf("No pool found with name_label %s", poolName),
			)
			return
		} else if len(pools) > 1 {
			resp.Diagnostics.AddError(
				"Multiple pools found",
				fmt.Sprintf("Multiple pools found with name_label %s. Please provide a unique name_label or use the `id` attribute to specify the pool.", poolName),
			)
			return
		}
		pool = pools[0]
	} else if data.NameLabel.IsUnknown() || data.NameLabel.IsNull() {
		// Retrieve pool by id if name_label is not provided
		poolID, err := uuid.FromString(data.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Invalid pool ID",
				fmt.Sprintf("The provided pool ID is not a valid UUID: %s", data.ID.ValueString()),
			)
			return
		}

		pool, err = d.client.Pool().Get(ctx, poolID)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error reading pool",
				fmt.Sprintf("Could not read pool %s: %s", poolID, err.Error()),
			)
			return
		}
	} else {
		resp.Diagnostics.AddError(
			"Invalid pool data source configuration",
			"Exactly one of `id` or `name_label` must be provided.",
		)
		return
	}

	result, diags := poolModelFromPayload(ctx, pool)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}
