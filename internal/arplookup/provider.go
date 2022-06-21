package arplookup

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"inet.af/netaddr"
)

var _ tfsdk.Provider = &provider{}

type provider struct {
	configured bool
	version    string
	network    netaddr.IPSet
	timeout    time.Duration
}

type providerData struct {
	Network types.List   `tfsdk:"network"`
	Timeout types.String `tfsdk:"timeout"`
}

func (p *provider) Configure(ctx context.Context, req tfsdk.ConfigureProviderRequest, resp *tfsdk.ConfigureProviderResponse) {
	var data providerData
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	p.network = netaddr.IPSet{}
	if !data.Network.Null {
		networks := []string{}
		data.Network.ElementsAs(ctx, &networks, false)
		network, err := mkIPSet(networks)
		if err != nil {
			resp.Diagnostics.AddError("unable to parse network CIDRs", err.Error())
			return
		}
		p.network = *network
	}

	if !data.Timeout.Null {
		timeout, err := time.ParseDuration(data.Timeout.Value)
		if err != nil {
			resp.Diagnostics.AddError("unable to parse timeout", err.Error())
			return
		}
		p.timeout = timeout
	}

	p.configured = true
}

func (p *provider) GetResources(ctx context.Context) (map[string]tfsdk.ResourceType, diag.Diagnostics) {
	return map[string]tfsdk.ResourceType{}, nil
}

func (p *provider) GetDataSources(ctx context.Context) (map[string]tfsdk.DataSourceType, diag.Diagnostics) {
	return map[string]tfsdk.DataSourceType{
		"arplookup_ip": ipDataSourceType{},
	}, nil
}

func (p *provider) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"network": {
				MarkdownDescription: "Network CIDR to search for.",
				Optional:            true,
				Type: types.ListType{
					ElemType: types.StringType,
				},
				PlanModifiers: tfsdk.AttributePlanModifiers{
					tfsdk.RequiresReplace(),
				},
			},
			"timeout": {
				MarkdownDescription: "Timeout for ARP lookup.",
				Optional:            true,
				Type:                types.StringType,
			},
		},
	}, nil
}

func New(version string) func() tfsdk.Provider {
	return func() tfsdk.Provider {
		return &provider{
			version: version,
		}
	}
}

func convertProviderType(in tfsdk.Provider) (provider, diag.Diagnostics) {
	var diags diag.Diagnostics
	p, ok := in.(*provider)
	if !ok {
		diags.AddError(
			"Unexpected Provider Instance Type",
			fmt.Sprintf("While creating the data source or resource, an unexpected provider type (%T) was received. This is always a bug in the provider code and should be reported to the provider developers.", p),
		)
		return provider{}, diags
	}

	if p == nil {
		diags.AddError(
			"Unexpected Provider Instance Type",
			"While creating the data source or resource, an unexpected empty provider instance was received. This is always a bug in the provider code and should be reported to the provider developers.",
		)
		return provider{}, diags
	}

	return *p, diags
}
