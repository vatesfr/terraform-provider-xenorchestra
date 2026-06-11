package data_sources

import (
	"context"
	"fmt"

	"github.com/gofrs/uuid"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vatesfr/terraform-provider-xenorchestra/provider/helpers"
	"github.com/vatesfr/xenorchestra-go-sdk/pkg/payloads"
	v2 "github.com/vatesfr/xenorchestra-go-sdk/v2"
)

var _ datasource.DataSource = &PoolDataSource{}

type PoolDataSource struct {
	client *v2.XOClient
}

type PoolDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	NameLabel   types.String `tfsdk:"name_label"`
	Description types.String `tfsdk:"description"`
	Master      types.String `tfsdk:"master"`
	DefaultSR   types.String `tfsdk:"default_sr"`
	Cores       types.Int64  `tfsdk:"cores"`
	Sockets     types.Int64  `tfsdk:"sockets"`
	Tags        types.List   `tfsdk:"tags"`
}

func poolSchemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "The id of the pool.",
			Computed:    true,
		},
		"name_label": schema.StringAttribute{
			Description: "The name label of the pool.",
			Computed:    true,
		},
		"description": schema.StringAttribute{
			Description: "The description of the pool.",
			Computed:    true,
		},
		"master": schema.StringAttribute{
			Description: "The id of the primary instance in the pool.",
			Computed:    true,
		},
		"default_sr": schema.StringAttribute{
			Description: "The default storage repository for the pool.",
			Computed:    true,
		},
		"cores": schema.Int64Attribute{
			Description: "Number of CPU cores in the pool.",
			Computed:    true,
		},
		"sockets": schema.Int64Attribute{
			Description: "Number of CPU sockets in the pool.",
			Computed:    true,
		},
		"tags": schema.ListAttribute{
			Description: "The tags associated with the pool.",
			Computed:    true,
			ElementType: types.StringType,
		},
	}
}

func poolModelFromPayload(ctx context.Context, pool *payloads.Pool) (PoolDataSourceModel, diag.Diagnostics) {
	tags, diags := types.ListValueFrom(ctx, types.StringType, pool.Tags)

	result := PoolDataSourceModel{
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

func NewPoolDataSource() datasource.DataSource {
	return &PoolDataSource{}
}

func (d *PoolDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_pool"
}

func (d *PoolDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {

	// Set "id" to required instead of computed for the single pool data source
	attrs := poolSchemaAttributes()
	attrs["id"] = schema.StringAttribute{
		Description: "The id of the pool.",
		Required:    true,
	}

	resp.Schema = schema.Schema{
		Description: "Use this data source to get information about a Xen Orchestra pool.",
		Attributes:  attrs,
	}
}

func (d *PoolDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = helpers.ProviderDataToXOClient(req.ProviderData, &resp.Diagnostics)
}

func (d *PoolDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data PoolDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	poolID, err := uuid.FromString(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid pool ID",
			fmt.Sprintf("The provided pool ID is not a valid UUID: %s", data.ID.ValueString()),
		)
		return
	}

	pool, err := d.client.Pool().Get(ctx, poolID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading pool",
			fmt.Sprintf("Could not read pool %s: %s", poolID, err.Error()),
		)
		return
	}

	result, diags := poolModelFromPayload(ctx, pool)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		resp.Diagnostics.AddError(
			"Error converting pool tags",
			"Could not convert pool tags to Terraform types.",
		)
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}
