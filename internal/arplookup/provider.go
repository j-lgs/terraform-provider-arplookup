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
				Validators: []tfsdk.AttributeValidator{
					networkValidator{},
				},
			},
			"timeout": {
				MarkdownDescription: `Timeout for ARP lookup.
Global attribute that can be overidden by being set in data sources.`,
				Optional: true,
				Type:     types.StringType,
				Validators: []tfsdk.AttributeValidator{
					timeValidator{},
				},
			},
			"backoff": {
				MarkdownDescription: `How long to wait between scans of the IP ranges specified by ` + "`network`" + `.
Global attribute that can be overidden by being set in data sources.`,
				Optional: true,
				Type:     types.StringType,
				Validators: []tfsdk.AttributeValidator{
					timeValidator{},
				},
			},
		},
	}, nil
}

type provider struct {
	configured bool
	version    string
	network    netaddr.IPSet
	timeout    time.Duration
	backoff    time.Duration
}

type providerData struct {
	Network types.List   `tfsdk:"network"`
	Timeout types.String `tfsdk:"timeout"`
	Backoff types.String `tfsdk:"backoff"`
}

func (data *providerData) configure(ctx context.Context, p *provider) error {
	p.network = netaddr.IPSet{}
	p.timeout = 5 * time.Minute
	p.backoff = 5 * time.Second

	if !data.Network.Null {
		networks := []string{}
		data.Network.ElementsAs(ctx, &networks, false)
		network, err := mkIPSet(networks)
		if err != nil {
			return err
		}
		p.network = *network
	}

	if !data.Timeout.Null {
		timeout, err := time.ParseDuration(data.Timeout.Value)
		if err != nil {
			return err
		}
		p.timeout = timeout
	}

	if !data.Backoff.Null {
		backoff, err := time.ParseDuration(data.Backoff.Value)
		if err != nil {
			return err
		}
		p.backoff = backoff
	}

	p.configured = true

	return nil
}

func (p *provider) Configure(ctx context.Context, req tfsdk.ConfigureProviderRequest, resp *tfsdk.ConfigureProviderResponse) {
	var data providerData
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := data.configure(ctx, p); err != nil {
		resp.Diagnostics.AddError("issue encountered running provider configure", err.Error())
		return
	}
}

func (p *provider) GetResources(ctx context.Context) (map[string]tfsdk.ResourceType, diag.Diagnostics) {
	return map[string]tfsdk.ResourceType{}, nil
}

func (p *provider) GetDataSources(ctx context.Context) (map[string]tfsdk.DataSourceType, diag.Diagnostics) {
	return map[string]tfsdk.DataSourceType{
		"arplookup_ip": ipDataSourceType{},
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
