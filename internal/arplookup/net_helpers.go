package arplookup

import (
	"net/netip"

	"inet.af/netaddr"
)

// toNetaddr is a convenience function to convert a netip.Addr to a netaddr.IP
func toNetaddr(ip netip.Addr) netaddr.IP {
	if ip.Is4() {
		return netaddr.IPFrom4(ip.As4())
	}

	return netaddr.IPFrom16(ip.As16())
}

// fromNetaddr is a convenience function to convert a netaddr.IP to a netip.Addr
func fromNetaddr(ip netaddr.IP) netip.Addr {
	if ip.Is4() {
		return netip.AddrFrom4(ip.As4())
	}

	return netip.AddrFrom16(ip.As16())
}

// isValidHost checks whether the host is not an IPv4 broadcast or unspecified address.
func isValidHost(ip netaddr.IP) bool {
	if ip.Is4() {
		bytes := ip.As4()
		return !ip.IsLoopback() && !ip.IsMulticast() && !ip.IsUnspecified() && ip.IsPrivate() && bytes[3] != 0xff && bytes[3] != 0x00
	}
	return false
}

// mkIPSet builds an IP set from a set of subnets in CIDR prefix notation.
func mkIPSet(networks []string) (*netaddr.IPSet, error) {
	var ips netaddr.IPSetBuilder

	for _, prefix := range networks {
		ip, err := netaddr.ParseIPPrefix(prefix)
		if err != nil {
			return nil, err
		}

		ips.AddPrefix(ip)
	}

	return ips.IPSet()
}
