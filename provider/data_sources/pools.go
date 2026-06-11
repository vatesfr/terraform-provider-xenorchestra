package data_sources

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vatesfr/terraform-provider-xenorchestra/provider/helpers"
	"github.com/vatesfr/xenorchestra-go-sdk/pkg/payloads"
	v2 "github.com/vatesfr/xenorchestra-go-sdk/v2"
)

var _ datasource.DataSource = &PoolsDataSource{}

type PoolsDataSource struct {
	client *v2.XOClient
}

type PoolsDataSourceModel struct {
	Tags      types.Set    `tfsdk:"tags"`
	SortBy    types.String `tfsdk:"sort_by"`
	SortOrder types.String `tfsdk:"sort_order"`
	Pools     types.List   `tfsdk:"pools"`
}

func NewPoolsDataSource() datasource.DataSource {
	return &PoolsDataSource{}
}

func (d *PoolsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_pools"
}

func (d *PoolsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to filter Xen Orchestra pools by certain criteria (tags) for use in other resources.",
		Attributes: map[string]schema.Attribute{
			"tags": schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "A list of tags to filter pools by. Only pools with ALL specified tags will be returned.",
			},
			"sort_by": schema.StringAttribute{
				Optional:    true,
				Description: "The pool field to sort the results by. Supported values: `id`, `name_label`.",
			},
			"sort_order": schema.StringAttribute{
				Optional:    true,
				Description: "The sort order. Supported values: `asc`, `desc`. Default is `asc`.",
			},
			"pools": schema.ListNestedAttribute{
				Computed:    true,
				Description: "The resulting pools after applying the argument filtering.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: poolSchemaAttributes(),
				},
			},
		},
	}
}

func (d *PoolsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = helpers.ProviderDataToXOClient(req.ProviderData, &resp.Diagnostics)
}

func (d *PoolsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data PoolsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var tags []string
	if !data.Tags.IsNull() && !data.Tags.IsUnknown() {
		resp.Diagnostics.Append(data.Tags.ElementsAs(ctx, &tags, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
	var tagsString string
	if len(tags) > 0 {
		tagsString = strings.Join(tags, " tags:") // Space separates multiple tags
	}

	var pools []*payloads.Pool
	pools, err := d.client.Pool().GetAll(ctx, 0, tagsString)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading pools",
			fmt.Sprintf("Could not list pools: %s", err.Error()),
		)
		return
	}

	sortBy := data.SortBy.ValueString()
	sortOrder := data.SortOrder.ValueString()

	if sortBy != "" {
		if sortOrder == "" {
			sortOrder = "asc"
		}

		if sortBy != "id" && sortBy != "name_label" {
			resp.Diagnostics.AddAttributeError(
				path.Root("sort_by"),
				"Invalid sort_by value",
				fmt.Sprintf("sort_by must be one of: id, name_label. Got: %s", sortBy),
			)
			return
		}

		sort.Slice(pools, func(i, j int) bool {
			var less bool
			switch sortBy {
			case "id":
				less = pools[i].ID.String() < pools[j].ID.String()
			case "name_label":
				less = pools[i].NameLabel < pools[j].NameLabel
			}

			if sortOrder == "desc" {
				return !less
			}
			return less
		})
	}

	poolsTf := make([]PoolDataSourceModel, 0, len(pools))
	for _, pool := range pools {
		poolTf, diags := poolModelFromPayload(ctx, pool)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		poolsTf = append(poolsTf, poolTf)
	}

	poolElemType := types.ObjectType{
		AttrTypes: helpers.AttrTypesFromSchemaAttributes(poolSchemaAttributes()),
	}

	poolList, diags := types.ListValueFrom(ctx, poolElemType, poolsTf)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	result := PoolsDataSourceModel{
		Tags:      data.Tags,
		SortBy:    data.SortBy,
		SortOrder: data.SortOrder,
		Pools:     poolList,
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}
