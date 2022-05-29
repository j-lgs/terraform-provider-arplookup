package arplookup

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"os"
	"testing"
)

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
					resource.TestCheckResourceAttr("data.arplookup_ip.test", "id", "ip"),
				),
			},
		},
	})
}

func macaddr() string {
	v, _ := os.LookupEnv("MAC")
	return v
}

var testAccIPDataSourceConfig = `
data "arplookup_ip" "test" {
  macaddr = "` + macaddr() + `"
  network = "172.18.0.0/24"
}
`
