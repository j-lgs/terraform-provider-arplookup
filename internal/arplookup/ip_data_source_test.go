package arplookup

import (
	"fmt"
	"regexp"
	"testing"

	"terraform-provider-arplookup/internal/arplookup/testdriver"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

var (
	index   = 5
	ip      = fmt.Sprintf("10.18.%d.18", index+1)
	network = fmt.Sprintf("%s/17", ip)
	mac     = "3e:50:6e:54:28:3d"
)

var driver *testdriver.Driver = &testdriver.Driver{}

// Test whether an IP is successfully derived from a MAC address
func TestAccIPDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					if err := driver.Init(); err != nil {
						t.Fatalf("unable to init test driver: %s", err.Error())
					}

					if err := driver.Needle(mac, network, index); err != nil {
						t.Fatalf("unable to insert needle into test haystack: %s", err.Error())
					}
				},
				Config: testAccIPDataSourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.arplookup_ip.test", "id", mac),
					resource.TestCheckResourceAttr("data.arplookup_ip.test", "ip", ip),
				),
			}, {
				PreConfig: func() {
					if err := driver.EnsureNo(mac); err != nil {
						t.Fatalf("unable to ensure mac doesn't exist: %s", err.Error())
					}
				},
				Config: testAccIPDataSourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.arplookup_ip.test", "id", mac),
					resource.TestCheckResourceAttr("data.arplookup_ip.test", "ip", ip),
				),
			},
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
				PreConfig: func() {
					if err := driver.EnsureNo(mac); err != nil {
						t.Fatalf("unable to ensure mac doesn't exist: %s", err.Error())
					}
				},
				Config:      testAccIPInvalidMAC,
				ExpectError: regexp.MustCompile("error: IP address corresponding to given MAC address not found in system ARP"),
				Check:       resource.ComposeAggregateTestCheckFunc(),
			},
		},
	})
}

// TODO add ipv6
var testAccIPDataSourceConfig = `
provider "arplookup" {
  timeout = "5s"
}

data "arplookup_ip" "test" {
  macaddr = "` + mac + `"
  network = [
    "10.18.0.0/17",
  ]
}
`

// TODO add ipv6
var testAccIPInvalidMAC = `
provider "arplookup" {
  timeout = "5s"
}

data "arplookup_ip" "test" {
  macaddr = "0b:de:ad:be:ef:0b"
  network = [
    "10.18.0.0/21",
    "10.18.8.0/23",
    "10.18.10.0/24"
  ]
}
`
