package arplookup

import (
	"fmt"
	"math/rand"
	"regexp"
	"testing"
	"time"

	"terraform-provider-arplookup/internal/arplookup/testdriver"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

var (
	index    = 5
	ip       = fmt.Sprintf("10.18.%d.18", index+1)
	network  = fmt.Sprintf("%s/17", ip)
	mac      = "3e:50:6e:54:28:3d"
	wrongmac = "0b:de:ad:be:ef:0b"
)

var driver *testdriver.Driver = &testdriver.Driver{}

// Test whether an IP is successfully derived from a MAC address
func TestAccIPDataSource(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().UTC().Unix()))

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					if err := driver.Init(r); err != nil {
						t.Fatalf("unable to init test driver: %s", err.Error())
					}

					if err := driver.EnsureNo(mac); err != nil {
						t.Fatalf("unable to ensure mac doesn't exist: %s", err.Error())
					}

					if err := driver.Needle(mac, ip, network, index); err != nil {
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

var testAccIPDataSourceConfig = `
provider "arplookup" {
  timeout = "10s"
}

data "arplookup_ip" "test" {
  interface = "br0"
  macaddr = "` + mac + `"
  network = [
    "10.18.0.0/21",
    "10.18.8.0/23",
    "10.18.10.0/24"
  ]
}
`

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
					if err := driver.EnsureNo(wrongmac); err != nil {
						t.Fatalf("unable to ensure mac doesn't exist: %s", err.Error())
					}
				},
				Config:      testAccIPInvalidMAC,
				ExpectError: regexp.MustCompile("error: IP address corresponding to given MAC"),
				Check:       resource.ComposeAggregateTestCheckFunc(),
			},
		},
	})
}

var testAccIPInvalidMAC = `
provider "arplookup" {
  timeout = "2s"
}

data "arplookup_ip" "test" {
  interface = "br0"
  macaddr = "` + wrongmac + `"
  network = [
    "10.18.6.0/24"
  ]
}
`

func TestAccIPDataSourceWrongInterface(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccDataSourceWrongInterface,
				ExpectError: regexp.MustCompile("error getting network interface"),
				Check:       resource.ComposeAggregateTestCheckFunc(),
			},
		},
	})
}

var testAccDataSourceWrongInterface = `
data "arplookup_ip" "test" {
  interface = "br1"
  macaddr = "` + mac + `"
  network = [
    "10.18.0.1/32",
  ]
}
`
