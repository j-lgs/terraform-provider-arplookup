package arplookup

import (
	"context"
	"net"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ tfsdk.DataSourceType = ipDataSourceType{}
var _ tfsdk.DataSource = ipDataSource{}

type ipDataSourceType struct{}

func (t ipDataSourceType) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Example data source",

		Attributes: map[string]tfsdk.Attribute{
			"macaddr": {
				MarkdownDescription: "MAC address to search for.",
				Optional:            true,
				Type:                types.StringType,
			},
			"network": {
				MarkdownDescription: "Network to search for macaddr in.",
				Optional:            true,
				Type:                types.StringType,
			},
			"ip": {
				MarkdownDescription: "Resultant IP address.",
				Computed:            true,
				Type:                types.StringType,
			},
			"id": {
				MarkdownDescription: "Example identifier",
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
	Network types.String `tfsdk:"network"`
	MACAddr types.String `tfsdk:"macaddr"`
	IP      types.String `tfsdk:"ip"`
	Id      types.String `tfsdk:"id"`
}

type ipDataSource struct {
	provider provider
}

func (ipDataSource ipDataSource) Read(ctx context.Context, req tfsdk.ReadDataSourceRequest, resp *tfsdk.ReadDataSourceResponse) {
	var data ipDataSourceData
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	mac, err := net.ParseMAC(data.MACAddr.Value)
	if err != nil {
		resp.Diagnostics.AddError("unable to parse MAC address", err.Error())
		return
	}

	if !reflect.DeepEqual(ipDataSource.provider.network, net.IPNet{}) && data.Network.Null {
		resp.Diagnostics.AddError("no networks specified", "`network` must be specified in either the provider or the datasource.")
		return
	}

	var network *net.IPNet
	if !reflect.DeepEqual(ipDataSource.provider.network, net.IPNet{}) {
		network = &ipDataSource.provider.network
	}

	if !data.Network.Null {
		_, network, err = net.ParseCIDR(data.Network.Value)
		if err != nil {
			resp.Diagnostics.AddError("unable to parse network CIDR", err.Error())
			return
		}
	}

	ip, err := checkARP(ctx, mac, *network)
	if err != nil {
		resp.Diagnostics.AddError("unable to check system ARP cache", err.Error())
		return
	}

	data.IP = types.String{Value: ip.String()}
	data.Id = types.String{Value: "ip"}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}
