package arplookup

import (
	"context"
	"fmt"
	"net"
	"time"

	"inet.af/netaddr"
)

// Default timeout is 360 seconds.
const arpFuncBackoff = 5 * time.Second
const arpRequestDeadline = 1000 * time.Microsecond

// errNoIP is an error used when an IP cannot be found from an associated MAC address.
var errNoIP error = fmt.Errorf("error: IP address corresponding to given MAC address not found in system ARP table")

// getIPFor is a wrapper for checkARPRun to abstract out OS specific components.
func getIPFor(ctx context.Context, MAC net.HardwareAddr, data ctxData) (netaddr.IP, error) {
	return checkARPRun(ctx, mkLinuxARP(MAC), data)
}

// arpClient is an interface that describes a platform agnostic way of performing an ARP lookup for a MAC address.
type arpClient interface {
	init(*net.Interface) error // init any resources needed to perform ARP requests
	destroy() error            // destroy any resources needed to perform ARP requests
	// send a request to an IP to determine whether its MAC matches one specified in the implementation structure
	request(netaddr.IP) (IP, error)
	try(channels)   // read the system's ARP cache to avoid an expensive `request` call
	cache(IP) error // add an IP to the system's ARP cache
}

type IP struct {
	cached bool
	netaddr.IP
}

// lookupIPRange sends a request to all IPs in ipRange to determine whether their MAC matches the MAC in ac.
func lookupIPRange(ctx context.Context, ac arpClient, ranges []netaddr.IPRange, chans channels) {
	for _, ipRange := range ranges {
		for current := ipRange.From(); current.Compare(ipRange.To()) <= 0; current = current.Next() {
			select {
			case <-chans.stop:
				return
			default:
			}

			if !isValidHost(current) {
				continue
			}

			result, err := ac.request(current)
			if err != nil {
				chans.errors <- err
				return
			}
			if !result.IsZero() {
				chans.results <- result
				return
			}
		}

	}
}

// if timeout is greater or equal to arpfuncbackoff our runtime is greatly increased
type ctxData struct {
	iface   *net.Interface
	network *netaddr.IPSet
	backoff time.Duration
}

type stopType struct{}

type channels struct {
	results chan IP
	errors  chan error
	stop    chan stopType
}

func makeChannels() channels {
	return channels{
		results: make(chan IP, 1),
		errors:  make(chan error, 1),
		stop:    make(chan stopType, 1),
	}
}

// checkARPRun searches an ARP table for a given MAC address in a platform agnostic way. It is important
// to check if the returned error is macNotFoundError to determine the difference between a failure in
// operation and a failure to find the mac in the system's table.
func checkARPRun(ctx context.Context, ac arpClient, data ctxData) (ip netaddr.IP, err error) {
	ac.init(data.iface)
	defer ac.destroy()

	chans := makeChannels()
	defer close(chans.stop)

outer:
	for {
		go ac.try(chans)
		go lookupIPRange(ctx, ac, data.network.Ranges(), chans)

		t := time.NewTimer(data.backoff)
		select {
		case <-ctx.Done():
			t.Stop()

			chans.stop <- struct{}{}
			return netaddr.IP{}, errNoIP
		case err = <-chans.errors:
			break outer
		case ip := <-chans.results:
			if err = ac.cache(ip); err != nil {
				break outer
			}
			return ip.IP, nil
		case <-t.C:
		}
	}

	chans.stop <- struct{}{}
	return netaddr.IP{}, err
}
