package arplookup

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestNetInterfaceValidate(t *testing.T) {
	v := interfaceValidator{}

	ctx := context.Background()

	testcases := []struct {
		iface  string
		expect string
	}{
		{
			iface:  "lo",
			expect: "",
		},
		{
			iface:  "lol",
			expect: "error getting network interface",
		},
	}

	for _, test := range testcases {
		var iface attr.Value
		diags := tfsdk.ValueFrom(ctx, test.iface, types.StringType, &iface)
		if diags.HasError() {
			t.Fatal("unable to marshal go value to terraform value")
		}

		req := tfsdk.ValidateAttributeRequest{
			AttributePath:   path.Root("interface"),
			AttributeConfig: iface,
			Config:          tfsdk.Config{},
		}
		resp := &tfsdk.ValidateAttributeResponse{
			Diagnostics: make(diag.Diagnostics, 0),
		}

		v.Validate(ctx, req, resp)
		if resp.Diagnostics.HasError() && test.expect == "" {
			t.Fatalf("validation failed: %s %s",
				resp.Diagnostics[len(resp.Diagnostics)-1].Summary(),
				resp.Diagnostics[len(resp.Diagnostics)-1].Detail())
		}
		if resp.Diagnostics.HasError() && test.expect != resp.Diagnostics[len(resp.Diagnostics)-1].Summary() {
			t.Fatalf("unexpected error recieved: want %s, got %s",
				test.expect,
				resp.Diagnostics[len(resp.Diagnostics)-1].Summary())
		}
	}
}

func TestMACValidate(t *testing.T) {
	v := macValidator{}

	ctx := context.Background()

	testcases := []struct {
		mac    string
		expect string
	}{
		{
			mac:    "00:00:00:00:00:00",
			expect: "",
		},
		{
			mac:    "xx:xx:xx:xx:xx:xx",
			expect: "malformed or invalid MAC",
		},
	}

	for _, test := range testcases {
		var mac attr.Value
		diags := tfsdk.ValueFrom(ctx, test.mac, types.StringType, &mac)
		if diags.HasError() {
			t.Fatal("unable to marshal go value to terraform value")
		}

		req := tfsdk.ValidateAttributeRequest{
			AttributePath:   path.Root("interface"),
			AttributeConfig: mac,
			Config:          tfsdk.Config{},
		}
		resp := &tfsdk.ValidateAttributeResponse{
			Diagnostics: make(diag.Diagnostics, 0),
		}

		v.Validate(ctx, req, resp)
		if resp.Diagnostics.HasError() && test.expect == "" {
			t.Fatalf("validation failed: %s %s",
				resp.Diagnostics[len(resp.Diagnostics)-1].Summary(),
				resp.Diagnostics[len(resp.Diagnostics)-1].Detail())
		}
		if resp.Diagnostics.HasError() && test.expect != resp.Diagnostics[len(resp.Diagnostics)-1].Summary() {
			t.Fatalf("unexpected error recieved: want %s, got %s",
				test.expect,
				resp.Diagnostics[len(resp.Diagnostics)-1].Summary())
		}
	}
}

func TestTimeValidate(t *testing.T) {
	v := timeValidator{}

	ctx := context.Background()

	testcases := []struct {
		duration string
		expect   string
	}{
		{
			duration: "1s",
			expect:   "",
		},
		{
			duration: "1p",
			expect:   "malformed or invalid duration",
		},
	}

	for _, test := range testcases {
		var duration attr.Value
		diags := tfsdk.ValueFrom(ctx, test.duration, types.StringType, &duration)
		if diags.HasError() {
			t.Fatal("unable to marshal go value to terraform value")
		}

		req := tfsdk.ValidateAttributeRequest{
			AttributePath:   path.Root("interface"),
			AttributeConfig: duration,
			Config:          tfsdk.Config{},
		}
		resp := &tfsdk.ValidateAttributeResponse{
			Diagnostics: make(diag.Diagnostics, 0),
		}

		v.Validate(ctx, req, resp)
		if resp.Diagnostics.HasError() && test.expect == "" {
			t.Fatalf("validation failed: %s %s",
				resp.Diagnostics[len(resp.Diagnostics)-1].Summary(),
				resp.Diagnostics[len(resp.Diagnostics)-1].Detail())
		}
		if resp.Diagnostics.HasError() && test.expect != resp.Diagnostics[len(resp.Diagnostics)-1].Summary() {
			t.Fatalf("unexpected error recieved: want %s, got %s",
				test.expect,
				resp.Diagnostics[len(resp.Diagnostics)-1].Summary())
		}
	}
}

func TestNetworkValidate(t *testing.T) {
	v := networkValidator{}

	ctx := context.Background()

	testcases := []struct {
		name    string
		network []string
		expect  string
	}{
		{
			name:    "empty",
			network: []string{},
			expect:  "no networks specified",
		},
		{
			name: "correct",
			network: []string{
				"10.0.0.10/24",
				"10.0.1.10/18",
			},
			expect: "",
		},
		{
			name: "incorrect prefix",
			network: []string{
				"10.0.1.10/18",
				"10.0.0.10",
			},
			expect: "malformed or invalid CIDR prefix",
		},
		{
			name: "invalid ipset",
			network: []string{
				"10.0.1.10/0",
				"10.0.0.10/16",
			},
			expect: "provided IP prefixes create an invalid set of IPs",
		},
	}

	for _, test := range testcases {
		var network attr.Value
		diags := tfsdk.ValueFrom(ctx, test.network, types.ListType{ElemType: types.StringType}, &network)
		if diags.HasError() {
			t.Fatalf("unable to marshal go value to terraform value\n%s",
				diags[len(diags)-1].Detail())
		}

		req := tfsdk.ValidateAttributeRequest{
			AttributePath:   path.Root("interface"),
			AttributeConfig: network,
			Config:          tfsdk.Config{},
		}
		resp := &tfsdk.ValidateAttributeResponse{
			Diagnostics: make(diag.Diagnostics, 0),
		}

		v.Validate(ctx, req, resp)
		if resp.Diagnostics.HasError() && test.expect == "" {
			t.Fatalf("(case: %s) validation failed: %s %s",
				test.name,
				resp.Diagnostics[len(resp.Diagnostics)-1].Summary(),
				resp.Diagnostics[len(resp.Diagnostics)-1].Detail())
		}
		if resp.Diagnostics.HasError() && test.expect != resp.Diagnostics[len(resp.Diagnostics)-1].Summary() {
			t.Fatalf("(case: %s) unexpected error recieved: want \"%s\", got \"%s\"",
				test.name,
				test.expect,
				resp.Diagnostics[len(resp.Diagnostics)-1].Summary())
		}
	}
}
