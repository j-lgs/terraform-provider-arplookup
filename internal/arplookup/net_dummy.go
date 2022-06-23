package arplookup

import (
	"net"

	"inet.af/netaddr"
)

// dummyARP is a stub arpClient for unit testing.
type dummyARP struct {
	needle netaddr.IP
}

// mkDummyARP constructs a dummyARP struct.
func mkDummyARP(needle netaddr.IP) *dummyARP {
	return &dummyARP{
		needle: needle,
	}
}

// request implements arpClient for dummyARP. This is a dummy implementation intended for testing and will
// return a "needle" IP once it has been requested.
func (ac *dummyARP) request(current netaddr.IP) (ip IP, err error) {
	if current == ac.needle {
		return IP{cached: false, IP: ac.needle}, nil
	}

	return IP{}, nil
}

// init implements arpClient for dummyARP. This is a stub.
func (ac *dummyARP) init(*net.Interface) error { return nil }

// destroy implements arpClient for dummyARP. This is a stub.
func (ac *dummyARP) destroy() error { return nil }

func (ac *dummyARP) cache(IP) error { return nil }

func (ac *dummyARP) try(chan<- IP, chan<- error) {}
