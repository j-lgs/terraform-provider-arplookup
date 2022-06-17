package arplookup

import (
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// Test whether an IP is successfully derived from a MAC address
func TestAccIPDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccIPDataSourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.arplookup_ip.test", "id", macaddr()),
					resource.TestCheckResourceAttr("data.arplookup_ip.test", "ip", ip()),
				),
			},
			// Next step should switch the IP to a different subnet then test again, confirming we get the old variable.
			// This is expected behabiour as this is a oneshot resource. Making it dynamic would break a lot of guarentees
			// about terraform state and would be too "dynamic" for it's purpose (provisioning virtual machines in a DHCP network
			// environment).

			// Next step. Move test driver code to Go so it can be dynamically changed at test time.
			// Driver code will be called during the PreConfig() phase.
		},
	})
}

// Test whether changing the network recreates the resource

// Test that being created with an incorrect mac (or host that is down) results in failure after the timeout expires.
func TestAccIPDataSourceFails(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccIPInvalidMAC,
				ExpectError: regexp.MustCompile("error: IP address corresponding to given MAC address not found in system ARP"),
				Check:       resource.ComposeAggregateTestCheckFunc(),
			},
			// Next step should switch the IP to a different subnet then test again, confirming we get the old variable.
			// This is expected behabiour as this is a oneshot resource. Making it dynamic would break a lot of guarentees
			// about terraform state and would be too "dynamic" for it's purpose (provisioning virtual machines in a DHCP network
			// environment).

			// Next step. Move test driver code to Go so it can be dynamically changed at test time.
			// Driver code will be called during the PreConfig() phase.
		},
	})
}

func macaddr() string {
	v, _ := os.LookupEnv("MAC")
	return v
}

func ip() string {
	v, _ := os.LookupEnv("IP")
	return v
}

var testAccIPDataSourceConfig = `
data "arplookup_ip" "test" {
  macaddr = "` + macaddr() + `"
  network = "172.18.0.0/24"
}
`

var testAccIPInvalidMAC = `
data "arplookup_ip" "test" {
  macaddr = "0b:de:ad:be:ef:0b"
  network = "172.18.0.0/24"
}
`
