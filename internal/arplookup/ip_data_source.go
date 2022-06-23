package arplookup

import (
	"context"
	"fmt"
	"net"
	"reflect"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"inet.af/netaddr"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ tfsdk.DataSourceType = ipDataSourceType{}
var _ tfsdk.DataSource = ipDataSource{}

type ipDataSourceType struct{}

func (t ipDataSourceType) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: "This data source will search `network` for a host matching `macaddr`. ",
		Attributes: map[string]tfsdk.Attribute{
			"macaddr": {
				MarkdownDescription: "MAC address to search for.",
				Optional:            true,
				Type:                types.StringType,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					tfsdk.RequiresReplace(),
				},
				Validators: []tfsdk.AttributeValidator{
					macValidator{},
				},
			},
			"backoff": {
				MarkdownDescription: "How long to wait between scans of the IP ranges specified by `network`.",
				Optional:            true,
				Type:                types.StringType,
				Validators: []tfsdk.AttributeValidator{
					timeValidator{},
				},
			},
			"network": {
				MarkdownDescription: "Network to search for macaddr in.",
				Optional:            true,
				Type: types.ListType{
					ElemType: types.StringType,
				},
				PlanModifiers: tfsdk.AttributePlanModifiers{
					tfsdk.RequiresReplace(),
				},
				Validators: []tfsdk.AttributeValidator{
					networkValidator{},
				},
			},
			"interface": {
				MarkdownDescription: "Interface to bind to when searching for machines.",
				Required:            true,
				Type:                types.StringType,
				Validators: []tfsdk.AttributeValidator{
					interfaceValidator{},
				},
			},
			"ip": {
				MarkdownDescription: "Resultant IP address.",
				Computed:            true,
				Type:                types.StringType,
			},
			"id": {
				MarkdownDescription: "Unique identifier.",
				Type:                types.StringType,
				Computed:            true,
			},
		},
	}, nil
}

func (t ipDataSourceType) NewDataSource(ctx context.Context, in tfsdk.Provider) (tfsdk.DataSource, diag.Diagnostics) {
	provider, diags := convertProviderType(in)
	return ipDataSource{
		provider: provider,
	}, diags
}

type ipDataSourceData struct {
	Backoff   types.String `tfsdk:"backoff"`
	Network   types.List   `tfsdk:"network"`
	MACAddr   types.String `tfsdk:"macaddr"`
	Interface types.String `tfsdk:"interface"`
	IP        types.String `tfsdk:"ip"`
	Id        types.String `tfsdk:"id"`
}

type ipDataSource struct {
	provider provider
}

func (data *ipDataSourceData) read(ctx context.Context, ipDataSource ipDataSource) error {
	mac, err := net.ParseMAC(data.MACAddr.Value)
	if err != nil {
		return err
	}

	if reflect.DeepEqual(ipDataSource.provider.network, net.IPNet{}) && data.Network.Null {
		return fmt.Errorf("neither network specified")
	}

	var network *netaddr.IPSet
	if !reflect.DeepEqual(ipDataSource.provider.network, net.IPNet{}) {
		network = &ipDataSource.provider.network
	}

	if !data.Network.Null {
		networks := []string{}
		data.Network.ElementsAs(ctx, &networks, false)

		network, err = mkIPSet(networks)
		if err != nil {
			return err
		}
	}

	backoff := ipDataSource.provider.backoff
	if !data.Backoff.Null {
		backoff, err = time.ParseDuration(data.Backoff.Value)
		if err != nil {
			return err
		}
	}

	iface, err := net.InterfaceByName(data.Interface.Value)
	if err != nil {
		return err
	}

	ip, err := getIPFor(ctx, mac, ctxData{iface: iface, network: network, backoff: backoff})
	if err != nil {
		return fmt.Errorf("error running getIPFor: %w", err)
	}

	data.IP = types.String{Value: ip.String()}
	data.Id = types.String{Value: mac.String()}

	return nil
}

func (ipDataSource ipDataSource) Read(ctx context.Context, req tfsdk.ReadDataSourceRequest, resp *tfsdk.ReadDataSourceResponse) {
	var data ipDataSourceData
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, ipDataSource.provider.timeout)
	defer cancel()

	if err := data.read(ctx, ipDataSource); err != nil {
		resp.Diagnostics.AddError("issue encountered while looking up IP", err.Error())
		return
	}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}
