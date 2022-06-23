package arplookup

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"inet.af/netaddr"
)

// interfaceValidator validates whether a given network interface exists on the host.
type interfaceValidator struct{}

// Description implements AttributeValidator.
func (v interfaceValidator) Description(context.Context) string {
	return "Checks whether a valid and existing network interface has been passed to the provider."
}

// MarkdownDescription implements AttributeValidator.
func (v interfaceValidator) MarkdownDescription(context.Context) string {
	return "Checks whether a valid and existing network interface has been passed to the provider."
}

// Validate implements AttributeValidator.
func (v interfaceValidator) Validate(ctx context.Context, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse) {
	interfaceValue, err := req.AttributeConfig.ToTerraformValue(ctx)
	if err != nil {
		resp.Diagnostics.AddError("error converting attribute value to terraform value", err.Error())
		return
	}

	var iface string
	if err := interfaceValue.As(&iface); err != nil {
		resp.Diagnostics.AddError("error converting terraform value to go value", err.Error())
		return
	}

	_, err = net.InterfaceByName(iface)
	if err != nil {
		resp.Diagnostics.AddAttributeError(req.AttributePath, fmt.Sprintf("error getting network interface \"%s\" by name", iface), err.Error())
		return
	}
}

// macValidator checks whether a given MAC address is properly formed.
type macValidator struct{}

// Description implements AttributeValidator.
func (v macValidator) Description(context.Context) string {
	return "Checks whether a valid MAC address has been passed to the provider."
}

// MarkdownDescription implements AttributeValidator.
func (v macValidator) MarkdownDescription(context.Context) string {
	return "Checks whether a valid MAC address has been passed to the provider."
}

// Validate implements AttributeValidator.
func (v macValidator) Validate(ctx context.Context, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse) {
	macValue, err := req.AttributeConfig.ToTerraformValue(ctx)
	if err != nil {
		resp.Diagnostics.AddError("error converting attribute value to terraform value", err.Error())
		return
	}

	var mac string
	if err := macValue.As(&mac); err != nil {
		resp.Diagnostics.AddError("error converting terraform value to go value", err.Error())
		return
	}

	_, err = net.ParseMAC(mac)
	if err != nil {
		resp.Diagnostics.AddAttributeError(req.AttributePath, fmt.Sprintf("malformed or invalid MAC \"%s\" provided", mac), err.Error())
		return
	}
}

// timeValidator checks whether a given string representing a duration is a valid go duration.
type timeValidator struct{}

// Description implements AttributeValidator.
func (v timeValidator) Description(context.Context) string {
	return "Checks whether a valid go duration has been passed to the provider."
}

// MarkdownDescription implements AttributeValidator.
func (v timeValidator) MarkdownDescription(context.Context) string {
	return "Checks whether a valid go duration has been passed to the provider."
}

// Validate implements AttributeValidator.
func (v timeValidator) Validate(ctx context.Context, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse) {
	durationValue, err := req.AttributeConfig.ToTerraformValue(ctx)
	if err != nil {
		resp.Diagnostics.AddError("error converting attribute value to terraform value", err.Error())
		return
	}

	if durationValue.IsNull() {
		return
	}

	var duration string
	if err := durationValue.As(&duration); err != nil {
		resp.Diagnostics.AddError("error converting terraform value to go value", err.Error())
		return
	}

	_, err = time.ParseDuration(duration)
	if err != nil {
		resp.Diagnostics.AddAttributeError(req.AttributePath, fmt.Sprintf("malformed or invalid duration \"%s\" provided", duration), err.Error())
		return
	}
}

// networkValidator checks whether an attribute containing a list of CIDR prefixed (as strings) represents a valid netaddr.IPSet
type networkValidator struct{}

// Description implements AttributeValidator.
func (v networkValidator) Description(context.Context) string {
	return "Checks whether the attribute represents a valid netaddr.IPSet."
}

// MarkdownDescription implements AttributeValidator.
func (v networkValidator) MarkdownDescription(context.Context) string {
	return "Checks whether the attribute represents a valid `netaddr.IPSet`."
}

// Validate implements AttributeValidator.
func (v networkValidator) Validate(ctx context.Context, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse) {
	networkValue, err := req.AttributeConfig.ToTerraformValue(ctx)
	if err != nil {
		resp.Diagnostics.AddError("error converting attribute value to terraform value", err.Error())
		return
	}

	if networkValue.IsNull() {
		return
	}

	var builder netaddr.IPSetBuilder

	var networkValues []tftypes.Value
	if err := networkValue.As(&networkValues); err != nil {
		resp.Diagnostics.AddError("error converting terraform value to go value", err.Error())
		return
	}

	var networks = make([]string, len(networkValues))
	for i, value := range networkValues {
		value.As(&networks[i])
	}

	for _, network := range networks {
		p, err := netaddr.ParseIPPrefix(network)
		if err != nil {
			resp.Diagnostics.AddAttributeError(req.AttributePath,
				fmt.Sprintf("malformed or invalid CIDR prefix \"%s\" provided", network), err.Error())
			return
		}

		builder.AddPrefix(p)
	}

	_, err = builder.IPSet()
	if err != nil {
		resp.Diagnostics.AddAttributeError(req.AttributePath, "provided IP prefixes create an invalid set of IPs", err.Error())
		return
	}
}
