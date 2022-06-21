package arplookup

import (
	"context"
	"fmt"
	"net"
	"net/netip"
	"time"

	"inet.af/netaddr"
)

func toNetaddr(ip netip.Addr) netaddr.IP {
	if ip.Is4() {
		return netaddr.IPFrom4(ip.As4())
	}

	return netaddr.IPFrom16(ip.As16())
}

func fromNetaddr(ip netaddr.IP) netip.Addr {
	if ip.Is4() {
		return netip.AddrFrom4(ip.As4())
	}

	return netip.AddrFrom16(ip.As16())
}

type arpClient interface {
	init() error
	request(netaddr.IP) (netaddr.IP, error)
	destroy() error
}

func doArp(ac arpClient, ipRange netaddr.IPRange) (netaddr.IP, error) {
	// check all IPs in range, attempting to find correct MAC address
	for current := ipRange.From(); current.Compare(ipRange.To()) <= 0; current = current.Next() {
		if !isValidHost(current) {
			continue
		}

		ip, err := ac.request(current)
		if err != nil {
			return netaddr.IP{}, err
		}
		if !ip.IsZero() {
			return ip, nil
		}
	}

	return netaddr.IP{}, nil
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

// Default timeout is 360 seconds.
const arpFuncTimeout = 360 * time.Second
const arpFuncBackoff = 5 * time.Second
const arpRequestDeadline = 100 * time.Microsecond

// errNoIP is an error used when an IP cannot be found from an associated MAC address.
var errNoIP error = fmt.Errorf("error: IP address corresponding to given MAC address not found in system ARP table")

// checkARP is a wrapper for checkARPRun to abstract out OS specific components.
func checkARP(ctx context.Context, MAC net.HardwareAddr, network netaddr.IPSet) (netaddr.IP, error) {
	return checkARPRun(ctx, network, mkLinuxARP(MAC))
}

// checkARPRun searches an ARP table for a given MAC address in a platform agnostic way. It is important
// to check if the returned error is macNotFoundError to determine the difference between a failure in
// operation and a failure to find the mac in the system's table.
func checkARPRun(ctx context.Context, network netaddr.IPSet, ac arpClient) (ip netaddr.IP, err error) {
	ac.init()

	ip, err = doArp(ac, network.Ranges()[0])
	if err != nil {
		return netaddr.IP{}, err
	}
	if !ip.IsZero() {
		return ip, nil
	}

	ctx, done := context.WithTimeout(ctx, arpFuncTimeout)
	defer done()
	defer ac.destroy()
	for {
		select {
		case <-ctx.Done():
			return netaddr.IP{}, errNoIP
		case <-time.After(arpFuncBackoff):
			ip, err = doArp(ac, network.Ranges()[0])
			if err != nil {
				return netaddr.IP{}, err
			}
			if !ip.IsZero() {
				return ip, nil
			}
		}
	}
}
