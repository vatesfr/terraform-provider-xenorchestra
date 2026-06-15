// Copyright (c) Vates
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/gofrs/uuid"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vatesfr/terraform-provider-xenorchestra/v2/internal/helpers"
	"github.com/vatesfr/xenorchestra-go-sdk/client"
	v2 "github.com/vatesfr/xenorchestra-go-sdk/v2"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &networkResource{}
var _ resource.ResourceWithImportState = &networkResource{}

var netDefaultDesc string = "Created with Xen Orchestra"

func NewNetworkResource() resource.Resource {
	return &networkResource{}
}

// networkResource defines the resource implementation.
type networkResource struct {
	client *v2.XOClient
}

// networkResourceModel describes the resource data model.
type networkResourceModel struct {
	ID              types.String `tfsdk:"id"`
	NameLabel       types.String `tfsdk:"name_label"`
	NameDescription types.String `tfsdk:"name_description"`
	Automatic       types.Bool   `tfsdk:"automatic"`
	DefaultIsLocked types.Bool   `tfsdk:"default_is_locked"`
	PIFDevice       types.String `tfsdk:"source_pif_device"`
	VLAN            types.Int64  `tfsdk:"vlan"`
	PoolID          types.String `tfsdk:"pool_id"`
	MTU             types.Int64  `tfsdk:"mtu"`
	Nbd             types.Bool   `tfsdk:"nbd"`
	InsecureNbd     types.Bool   `tfsdk:"insecure_nbd"`
	Tags            types.List   `tfsdk:"tags"`
}

func networkModelFromPayload(ctx context.Context, c *v2.XOClient, net client.Network) (networkResourceModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	pif, d := getPIFForNetwork(c, net)
	diags.Append(d...)
	if diags.HasError() {
		return networkResourceModel{}, diags
	}

	// TODO: use the v2 client once it supports network creation.
	// tagList := make([]string, 0)
	// tags, diags := types.ListValueFrom(ctx, types.StringType, tagList) //TODO: use actual tag list

	networkModel := networkResourceModel{
		ID:              types.StringValue(net.Id),
		NameLabel:       types.StringValue(net.NameLabel),
		NameDescription: types.StringValue(net.NameDescription),
		Automatic:       types.BoolValue(net.Automatic),
		DefaultIsLocked: types.BoolValue(net.DefaultIsLocked),
		// PIF:             types.StringValue(pifID),
		// VLAN:            types.Int64Value(pifVlan),
		PoolID:      types.StringValue(net.PoolId),
		MTU:         types.Int64Value(int64(net.MTU)),
		Nbd:         types.BoolValue(net.Nbd),
		InsecureNbd: types.BoolValue(net.InsecureNbd),
		Tags:        types.ListNull(types.StringType),
	}
	if pif != nil {
		networkModel.PIFDevice = types.StringValue(pif.Device)
		networkModel.VLAN = types.Int64Value(int64(pif.Vlan))
	}

	return networkModel, diags
}

func (r *networkResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_network"
}

func ptr[T any](v T) *T { return &v }

func (r *networkResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Network resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"automatic": schema.BoolAttribute{
				MarkdownDescription: "",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"name_label": schema.StringAttribute{
				MarkdownDescription: "The name label of the network.",
				Required:            true,
			},
			"name_description": schema.StringAttribute{
				MarkdownDescription: "The name description of the network.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(netDefaultDesc),
			},
			"default_is_locked": schema.BoolAttribute{
				MarkdownDescription: "This argument controls whether the network should enforce VIF locking. This defaults to `false` which means that no filtering rules are applied.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"source_pif_device": schema.StringAttribute{
				MarkdownDescription: "The PIF device (eth0, eth1, etc) that will be used as an input during network creation. This parameter is required if a vlan is specified. If not specified, it will create an internal network.",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.AlsoRequires(path.MatchRoot("vlan")),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"vlan": schema.Int64Attribute{
				MarkdownDescription: "The vlan to use for the network. Defaults to `0` meaning no VLAN.",
				Optional:            true,
				// Computed:            true,
				// Default:             int64default.StaticInt64(0),
				Validators: []validator.Int64{
					int64validator.AlsoRequires(path.MatchRoot("source_pif_device")),
				},
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"pool_id": schema.StringAttribute{
				MarkdownDescription: "The pool id that this network should belong to.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"mtu": schema.Int64Attribute{
				MarkdownDescription: "The MTU of the network. Defaults to `1500` if unspecified.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(1500),
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"nbd": schema.BoolAttribute{
				MarkdownDescription: "Whether the network should use a network block device. Defaults to `false` if unspecified.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"insecure_nbd": schema.BoolAttribute{
				MarkdownDescription: "Whether the network should use an insecure network block device. Defaults to `false` if unspecified.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"tags": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "The tags associated with the network.",
				Optional:            true,
			},
		},
	}
}

func (r *networkResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = helpers.ProviderDataToXOClient(req.ProviderData, &resp.Diagnostics)
}

func (r *networkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data networkResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	// httpResp, err := r.client.Do(httpReq)
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create example, got error: %s", err))
	//     return
	// }

	var pifID string
	if !data.PIFDevice.IsUnknown() && !data.PIFDevice.IsNull() {
		pif, err := getNetworkCreationSourcePIF(ctx, r.client, data.PIFDevice.ValueString(), data.PoolID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get PIF, got error: %s", err))
			return
		}
		pifID = pif.Id
	}

	// TODO: use the v2 client once it supports network creation.
	networkParams := client.CreateNetworkRequest{
		Automatic:       data.Automatic.ValueBool(),
		Name:            data.NameLabel.ValueString(),
		Description:     data.NameDescription.ValueString(),
		DefaultIsLocked: data.DefaultIsLocked.ValueBool(),
		PIF:             pifID,
		Vlan:            int(data.VLAN.ValueInt64()),
		Pool:            data.PoolID.ValueString(),
		Mtu:             int(data.MTU.ValueInt64()),
		Nbd:             data.Nbd.ValueBool(),
		// InsecureNbd:     data.InsecureNbd.ValueBool(),
	}
	network, err := r.client.V1Client().CreateNetwork(networkParams)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create network, got error: %s", err))
		return
	}

	result, diags := networkModelFromPayload(ctx, r.client, *network)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func (r *networkResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data networkResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	net, err := r.client.V1Client().GetNetwork(client.Network{Id: data.ID.ValueString()})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read network, got error: %s", err))
		return
	}

	result, diags := networkModelFromPayload(ctx, r.client, *net)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func (r *networkResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data networkResourceModel
	var plan networkResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// TODO: use the v2 client once it supports network updates.
	netUpdateReq := client.UpdateNetworkRequest{
		Id:              data.ID.ValueString(),
		Automatic:       plan.Automatic.ValueBool(),
		Nbd:             plan.Nbd.ValueBool(),
		NameLabel:       ptr(plan.NameLabel.ValueString()),
		NameDescription: ptr(plan.NameDescription.ValueString()),
		DefaultIsLocked: ptr(plan.DefaultIsLocked.ValueBool()),
	}

	net, err := r.client.V1Client().UpdateNetwork(netUpdateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update network, got error: %s", err))
		return
	}

	result, diags := networkModelFromPayload(ctx, r.client, *net)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func (r *networkResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data networkResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.V1Client().DeleteNetwork(data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete network, got error: %s", err))
		return
	}
}

func (r *networkResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// This function returns the PIF specified the given device name on the pool's primary host. In order to create
// networks with a VLAN, a PIF for the given device must be provided. Xen Orchestra uses the primary host's PIF
// for this and so we emulate that behavior.
func getNetworkCreationSourcePIF(ctx context.Context, c *v2.XOClient, pifDevice, poolId string) (*client.PIF, error) {
	pool, err := c.Pool().Get(ctx, uuid.Must(uuid.FromString(poolId)))
	if err != nil {
		return nil, err
	}

	// TODO: replace with proper v2 PIF service call once it is implemented.
	pifs, err := c.V1Client().GetPIF(client.PIF{
		Host:   pool.Master.String(),
		Vlan:   -1,
		Device: pifDevice,
	})

	if err != nil {
		return nil, err
	}

	if len(pifs) != 1 {
		return nil, fmt.Errorf("expected to find a single PIF, instead found %d. %+v", len(pifs), pifs)
	}

	return &pifs[0], nil
}

// Returns the VLAN and device name for the given network.
// TODO: replace with proper v2 PIF service call once it is implemented.
func getPIFForNetwork(c *v2.XOClient, net client.Network) (*client.PIF, diag.Diagnostics) {
	var diags diag.Diagnostics
	if len(net.PIFs) > 0 {
		pifs, err := c.V1Client().GetPIF(client.PIF{Id: net.PIFs[0]})
		if err != nil {
			diags.AddError(
				"v1 Client Error",
				fmt.Sprintf("Unable to get PIF for network, got error: %v", err),
			)
			return nil, diags
		}

		if len(pifs) != 1 {
			diags.AddError(
				"v1 Client Error",
				fmt.Sprintf("Expected to find a single PIF, instead found %d. %+v", len(pifs), pifs),
			)
			return nil, diags
		}
		return &pifs[0], nil
	}
	return nil, diags
}
