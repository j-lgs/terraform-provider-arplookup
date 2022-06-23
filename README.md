[![Tests](https://github.com/j-lgs/terraform-provider-arplookup/actions/workflows/test.yml/badge.svg)](https://github.com/j-lgs/terraform-provider-arplookup/actions/workflows/test.yml)

# terraform-provider-arplookup
A Terraform provider that contains a datasource which looks up an IP address in a network given an interface MAC address.

Check it out [Here](https://registry.terraform.io/providers/j-lgs/arplookup/latest)

# Use
Check the examples folder to see how the provider can be used. Also check out my [homelab provisioning](https://github.com/j-lgs/provisioning) repo to see the provider used to set up a Kubernetes cluster on Proxmox hosts.


Because the binary needs the NET_RAW capability (due to it's use of raw sockets) the following command must be ran after a `terraform init -upgrade`.
```
sudo setcap cap_net_raw,cap_net_admin=eip .terraform/providers/registry.terraform.io/j-lgs/arplookup/0.3.1/linux_amd64/terraform-provider-arplookup_v0.3.1
```

# Limitations
+ Has only been tested on my Linux system. Input, advice or PRs from Windows and MacOS users would be appreciated.
+ No testing with IPv6 has been done yet.

# Building
+ Install Go 1.18+
+ Clone the repo `git clone https://github.com/j-lgs/terraform-provider-arplookup.git`
+ Run unit tests with `make test`
+ Build the project with `make build`
+ Run acceptance tests with `make acctest`
