package arplookup

import (
	"context"
	"fmt"
	"net"
	"time"

	"inet.af/netaddr"
)

// arpClient is an interface that describes a platform agnostic way of performing an ARP lookup for a MAC address.
type arpClient interface {
	init(*net.Interface) error // init any resources needed to perform ARP requests
	destroy() error            // destroy any resources needed to perform ARP requests
	// send a request to an IP to determine whether its MAC matches one specified in the implementation structure
	request(netaddr.IP) (netaddr.IP, error)
}

// doArp sends a request to all IPs in ipRange to determine whether their MAC matches the MAC in ac.
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

// Default timeout is 360 seconds.
const arpFuncTimeout = 360 * time.Second
const arpFuncBackoff = 5 * time.Second
const arpRequestDeadline = 100 * time.Microsecond

// errNoIP is an error used when an IP cannot be found from an associated MAC address.
var errNoIP error = fmt.Errorf("error: IP address corresponding to given MAC address not found in system ARP table")

// checkARP is a wrapper for checkARPRun to abstract out OS specific components.
func checkARP(ctx context.Context, MAC net.HardwareAddr, network netaddr.IPSet, iface *net.Interface) (netaddr.IP, error) {
	return checkARPRun(ctx, network, mkLinuxARP(MAC), iface)
}

// checkARPRun searches an ARP table for a given MAC address in a platform agnostic way. It is important
// to check if the returned error is macNotFoundError to determine the difference between a failure in
// operation and a failure to find the mac in the system's table.
func checkARPRun(ctx context.Context, network netaddr.IPSet, ac arpClient, iface *net.Interface) (ip netaddr.IP, err error) {
	ac.init(iface)

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
